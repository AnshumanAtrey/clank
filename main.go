package main // This declares the package name. 'main' is the starting point of the program.

import (
	"flag"    // Importing the 'flag' package to handle command-line flags.
	"fmt"     // Importing the 'fmt' package for formatted I/O operations.
	"os"      // Importing the 'os' package to interact with the operating system.
	"strings" // Importing the 'strings' package for string manipulation functions.
	// New import for rate limiting
	// New import for HTTP requests
)

func main() { // The main function is the entry point of the program.
    // Define flags
    numberFlag := flag.String("n", "", "Phone number pattern (use x as placeholder)") // Define a string flag '-n' with a default value of an empty string and a description.
    flag.Parse() // Parse the command-line flags.

    var pattern string // Declare a variable to hold the phone number pattern.
    if *numberFlag == "" { // Check if the numberFlag is empty.
        // If no flag provided, check if there's a direct argument
        if len(os.Args) > 1 { // Check if there are command-line arguments provided.
            pattern = os.Args[1] // If so, assign the first argument to the pattern variable.
        } else { // If no argument is provided.
            fmt.Println("Usage: clank -n <phone_number_pattern>") // Print usage instructions.
            fmt.Println("Example: clank -n 918115605xxx") // Provide an example of how to use the program.
            os.Exit(1) // Exit the program with a status code of 1 (indicating an error).
        }
    } else { // If the numberFlag is not empty.
        pattern = *numberFlag // Assign the value of the numberFlag to the pattern variable.
    }

    // Validate input
    if !isValidPattern(pattern) { // Call the isValidPattern function to check if the pattern is valid.
        fmt.Println("Error: Pattern should only contain numbers and 'x' placeholders") // Print an error message if the pattern is invalid.
        os.Exit(1) // Exit the program with a status code of 1.
    }

    // Generate combinations
    combinations := generateCombinations(pattern) // Call the generateCombinations function to generate all possible combinations based on the pattern.

    // Print ASCII Art
    fmt.Println(
`________  ___       ________  _________  ___  ___       
|\   ____\|\  \     |\   __  \|\   ___  \|\  \|\  \     
\ \  \___|\ \  \    \ \  \|\  \ \  \\ \  \ \  \/  /|_   
 \ \  \    \ \  \    \ \   __  \ \  \\ \  \ \   ___  \  
  \ \  \____\ \  \____\ \  \ \  \ \  \\ \  \ \  \\ \  \ 
   \ \_______\ \_______\ \__\ \__\ \__\\ \__\ \__\\ \__\
    \|_______|\|_______|\|__|\|__|\|__| \|__|\|__| \|__|`)

    // Print results
    fmt.Printf("Generated %d combinations for pattern %s:\n", len(combinations), pattern) // Print the number of generated combinations and the pattern.
    for _, number := range combinations { // Loop through each generated number.
        fmt.Println(number) // Print each generated number.
    }
}

// Function to validate the phone number pattern
func isValidPattern(pattern string) bool { // Define a function that takes a string pattern and returns a boolean.
    for _, char := range pattern { // Loop through each character in the pattern.
        if char != 'x' && char != 'X' && (char < '0' || char > '9') { // Check if the character is not 'x' or a digit.
            return false // If an invalid character is found, return false.
        }
    }
    return true // If all characters are valid, return true.
}

// Function to generate all combinations based on the pattern
func generateCombinations(pattern string) []string { // Define a function that takes a string pattern and returns a slice of strings.
    // Convert pattern to lowercase for consistency
    pattern = strings.ToLower(pattern) // Convert the pattern to lowercase to handle 'x' and 'X' uniformly.
    
    // Count number of placeholders
    placeholders := strings.Count(pattern, "x") // Count how many 'x' characters are in the pattern.
    if placeholders == 0 { // If there are no placeholders.
        return []string{pattern} // Return a slice containing the original pattern.
    }

    // Calculate total combinations (10^placeholders)
    total := 1 // Initialize a variable to hold the total number of combinations.
    for i := 0; i < placeholders; i++ { // Loop for the number of placeholders.
        total *= 10 // Each placeholder can be replaced by 10 digits (0-9), so multiply total by 10.
    }

    combinations := make([]string, 0, total) // Create a slice to hold the generated combinations with an initial capacity of 'total'.
    
    // Generate all possible combinations
    for i := 0; i < total; i++ { // Loop through the total number of combinations.
        // Convert current number to a padded string
        replacement := fmt.Sprintf("%0*d", placeholders, i) // Format the current index 'i' as a zero-padded string based on the number of placeholders.
        
        // Create the new number by replacing x's with digits
        newNumber := pattern // Start with the original pattern.
        for _, digit := range replacement { // Loop through each digit in the replacement string.
            newNumber = strings.Replace(newNumber, "x", string(digit), 1) // Replace the first occurrence of 'x' in newNumber with the current digit.
        }
        
        combinations = append(combinations, newNumber) // Append the newly generated number to the combinations slice.
    }

    return combinations // Return the slice of generated combinations.
}