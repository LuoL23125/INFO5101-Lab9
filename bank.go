package main

import (
	"fmt"
	"sync"
	"time"
)

// Global variables for different tests
var balance int = 1000
var mu sync.Mutex

// Transaction represents a request from a customer account
type Transaction struct {
	Amount     int
	Source     string
	CustomerID string
	Type       string // "withdrawal" or "deposit"
}

// Ledger stores balances for both customer and bank
type Ledger struct {
	CustomerBalance int
	BankBalance     int
}

// ============================================
// PART 1: Race Condition Demo (from 1bank.go)
// ============================================
func withdrawRace(amount int, source string) {
	if balance >= amount {
		fmt.Printf("%s: is withdrawing %d\n", source, amount)
		time.Sleep(time.Millisecond * 500) // simulate processing time
		balance -= amount
		fmt.Printf("%s: Withdrawal of %d successful. New balance: %d\n", source, amount, balance)
	} else {
		fmt.Printf("%s: Withdrawal of %d failed: insufficient funds\n", source, amount)
	}
}

func testRaceCondition() {
	fmt.Println("\n========================================")
	fmt.Println("PART 1: RACE CONDITION (No Synchronization)")
	fmt.Println("========================================")
	balance = 1000
	fmt.Printf("Initial balance: %d\n\n", balance)
	
	go withdrawRace(700, "phone transaction")
	go withdrawRace(500, "atm transaction")
	
	time.Sleep(time.Second * 2) // wait for goroutines to finish
	
	fmt.Printf("\nFinal balance: %d", balance)
	fmt.Println(" ⚠️ INCORRECT - Race condition occurred!")
	fmt.Println("Expected: 300 or 500 (only one should succeed)")
	fmt.Printf("Got: %d (likely both succeeded, causing negative or wrong balance)\n", balance)
}

// ============================================
// PART 2: Channel-based (from 2bank.go with modifications)
// ============================================
func transactionProcessor(ledger *Ledger, txChan chan Transaction) {
	for tx := range txChan {
		// Display transaction type clearly
		if tx.Type == "deposit" {
			fmt.Printf("📥 Processing DEPOSIT: %s depositing %d for customer %s\n",
				tx.Source, tx.Amount, tx.CustomerID)
		} else {
			fmt.Printf("📤 Processing WITHDRAWAL: %s withdrawing %d for customer %s\n",
				tx.Source, tx.Amount, tx.CustomerID)
		}
		
		time.Sleep(time.Millisecond * 300) // simulate processing delay
		
		// Double-entry accounting based on transaction type
		if tx.Type == "withdrawal" {
			if ledger.CustomerBalance >= tx.Amount {
				ledger.CustomerBalance -= tx.Amount // Debit customer
				ledger.BankBalance += tx.Amount     // Credit bank
				fmt.Printf("✅ Withdrawal complete. Customer: %d | Bank: %d\n",
					ledger.CustomerBalance, ledger.BankBalance)
			} else {
				fmt.Printf("❌ Withdrawal failed: insufficient funds for %s\n", tx.CustomerID)
			}
		} else if tx.Type == "deposit" {
			ledger.CustomerBalance += tx.Amount // Credit customer
			ledger.BankBalance -= tx.Amount     // Debit bank
			fmt.Printf("✅ Deposit complete. Customer: %d | Bank: %d\n",
				ledger.CustomerBalance, ledger.BankBalance)
		}
	}
}

func testChannelBased() {
	fmt.Println("\n========================================")
	fmt.Println("PART 2: CHANNEL-BASED PROCESSING")
	fmt.Println("========================================")
	
	// Initial balances
	ledger := Ledger{
		CustomerBalance: 1000,
		BankBalance:     5000,
	}
	
	fmt.Printf("Initial - Customer: %d | Bank: %d\n\n", 
		ledger.CustomerBalance, ledger.BankBalance)
	
	// Channel for sending transaction requests
	txChan := make(chan Transaction)
	
	// Start the processor
	go transactionProcessor(&ledger, txChan)
	
	// Send original two transactions
	go func() {
		txChan <- Transaction{
			Amount:     700,
			Source:     "Phone Transfer",
			CustomerID: "CUST1001",
			Type:       "withdrawal",
		}
	}()
	
	go func() {
		txChan <- Transaction{
			Amount:     500,
			Source:     "ATM Withdrawal",
			CustomerID: "CUST1001",
			Type:       "withdrawal",
		}
	}()
	
	// BONUS: Third transaction - Online Purchase
	go func() {
		time.Sleep(time.Millisecond * 100) // slight delay to show sequence
		txChan <- Transaction{
			Amount:     400,
			Source:     "Online Purchase",
			CustomerID: "CUST1001",
			Type:       "withdrawal",
		}
	}()
	
	// BONUS: Deposit transaction - Salary Deposit
	go func() {
		time.Sleep(time.Millisecond * 200) // slight delay
		txChan <- Transaction{
			Amount:     1500,
			Source:     "Salary Deposit",
			CustomerID: "CUST1001",
			Type:       "deposit",
		}
	}()
	
	// Give goroutines time to send transactions
	time.Sleep(time.Second * 3)
	
	// Close channel
	close(txChan)
	time.Sleep(time.Millisecond * 100)
	
	fmt.Println("\nFinal Balances:")
	fmt.Println("Customer Balance:", ledger.CustomerBalance)
	fmt.Println("Bank Balance:", ledger.BankBalance)
	fmt.Printf("Total (Customer + Bank): %d (should equal 6000)\n",
		ledger.CustomerBalance+ledger.BankBalance)
}

// ============================================
// PART 3: Fixed with Mutex and WaitGroup
// ============================================
func withdrawMutex(amount int, source string, wg *sync.WaitGroup) {
	defer wg.Done()
	
	fmt.Printf("🔒 %s: attempting to withdraw %d\n", source, amount)
	
	mu.Lock()
	fmt.Printf("   %s: acquired lock\n", source)
	defer mu.Unlock()
	
	if balance >= amount {
		fmt.Printf("   %s: is withdrawing %d\n", source, amount)
		time.Sleep(time.Millisecond * 500) // simulate processing time
		balance -= amount
		fmt.Printf("   %s: Withdrawal of %d successful. New balance: %d\n", 
			source, amount, balance)
	} else {
		fmt.Printf("   %s: Withdrawal of %d failed: insufficient funds\n", 
			source, amount)
	}
}

// BONUS: Deposit function with mutex
func deposit(amount int, source string, wg *sync.WaitGroup) {
	defer wg.Done()
	
	fmt.Printf("💰 %s: depositing %d\n", source, amount)
	
	mu.Lock()
	fmt.Printf("   %s: acquired lock for deposit\n", source)
	defer mu.Unlock()
	
	time.Sleep(time.Millisecond * 300)
	balance += amount
	fmt.Printf("   %s: Deposit of %d successful. New balance: %d\n", 
		source, amount, balance)
}

func testMutex() {
	fmt.Println("\n========================================")
	fmt.Println("PART 3: FIXED WITH MUTEX & WAITGROUP")
	fmt.Println("========================================")
	balance = 1000
	fmt.Printf("Initial balance: %d\n\n", balance)
	
	var wg sync.WaitGroup
	
	// Original two transactions
	wg.Add(2)
	go withdrawMutex(700, "phone transaction", &wg)
	go withdrawMutex(500, "atm transaction", &wg)
	wg.Wait()
	
	fmt.Printf("\nBalance after two withdrawals: %d", balance)
	fmt.Println(" (300 or 500 depending on order)")
	
	// BONUS: Add more transactions including deposits
	fmt.Println("\n--- Adding more transactions (BONUS) ---")
	wg.Add(3)
	go withdrawMutex(400, "online purchase", &wg)      // Third withdrawal
	go deposit(1000, "salary deposit", &wg)            // Deposit
	go deposit(200, "refund", &wg)                     // Another deposit
	wg.Wait()
	
	fmt.Printf("\nFinal balance: %d ✅ CORRECT!\n", balance)
}

// ============================================
// BONUS: Comparison Test
// ============================================
func testComparison() {
	fmt.Println("\n========================================")
	fmt.Println("BONUS: COMPARISON - With vs Without Mutex")
	fmt.Println("========================================")
	
	// Test without mutex multiple times to show inconsistency
	fmt.Println("\n🔴 Testing WITHOUT Mutex (5 runs):")
	for i := 1; i <= 5; i++ {
		balance = 1000
		go withdrawRace(700, "phone")
		go withdrawRace(500, "atm")
		time.Sleep(time.Second)
		fmt.Printf("  Run %d: Final balance = %d\n", i, balance)
	}
	
	// Test with mutex to show consistency
	fmt.Println("\n🟢 Testing WITH Mutex (5 runs):")
	for i := 1; i <= 5; i++ {
		balance = 1000
		var wg sync.WaitGroup
		wg.Add(2)
		go withdrawMutex(700, "phone", &wg)
		go withdrawMutex(500, "atm", &wg)
		wg.Wait()
		fmt.Printf("  Run %d: Final balance = %d\n", i, balance)
	}
	
	fmt.Println("\n📊 OBSERVATIONS:")
	fmt.Println("Without Mutex:")
	fmt.Println("  • Both goroutines check balance=1000 simultaneously")
	fmt.Println("  • Both pass the if condition (1000>=700 and 1000>=500)")
	fmt.Println("  • Both execute withdrawal, causing incorrect result")
	fmt.Println("  • Final balance: -200 or unpredictable due to race")
	fmt.Println("\nWith Mutex:")
	fmt.Println("  • Only one goroutine can access balance at a time")
	fmt.Println("  • First transaction succeeds, second sees updated balance")
	fmt.Println("  • One succeeds, one fails due to insufficient funds")
	fmt.Println("  • Final balance: 300 or 500 (depending on which goes first)")
	fmt.Println("\nChannel Approach:")
	fmt.Println("  • Single processor handles all transactions sequentially")
	fmt.Println("  • Natural serialization prevents race conditions")
	fmt.Println("  • Implements proper double-entry bookkeeping")
}

// ============================================
// MAIN FUNCTION
// ============================================
func main() {
	fmt.Println("=====================================")
	fmt.Println("   GO MULTITHREADING BANKING LAB")
	fmt.Println("=====================================")
	
	// Run all tests in sequence
	testRaceCondition()     // Part 1: Shows the problem
	testChannelBased()      // Part 2: Channel solution
	testMutex()            // Part 3: Mutex solution
	testComparison()       // Bonus: Comparison
	
	fmt.Println("\n=====================================")
	fmt.Println("         ALL TESTS COMPLETE")
	fmt.Println("=====================================")
}