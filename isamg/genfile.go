package isamg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"isamg/julian"
	"unicode"
	"unicode/utf8"
)

func Insertion1(A *[]int) {
	for j := 1; j < len((*A)); j++ {
		i := j - 1
		key := (*A)[j]
		for i >= 0 && (*A)[i] > key {
			(*A)[i+1] = (*A)[i]
			i--
		}
		(*A)[i+1] = key
	}
}
func PrintArr1(A *[]int) {
	for i := 0; i < len(*A); i++ {
		fmt.Printf("A[%d]:%d", i, (*A)[i])
		fmt.Printf("\n")
	}
}
func loadFNames(strSlice *[]string) {

	file, err := os.Open("./isamg/data/fnames.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		*strSlice = append(*strSlice, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	//fmt.Println(lines)
}
func loadMNames(strSlice *[]string) {

	file, err := os.Open("./isamg/data/mnames.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		*strSlice = append(*strSlice, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	//fmt.Println(lines)
}
func formatName(name string) string {
	// Convert all to lowercase
	lowerName := strings.ToLower(name)
	r, size := utf8.DecodeRuneInString(lowerName)
	if r == utf8.RuneError {
		// handle error
	}
	return string(unicode.ToUpper(r)) + lowerName[size:]
}

// Return first n characters of a name in lower case
func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return strings.ToLower(s[:j])
		}
		i++
	}
	return s
}
func removeDashes(phoneNumber string) string {
	return strings.ReplaceAll(phoneNumber, "-", "")
}
func GenFile(dateStmp string) {
	var mnames []string
	var fnames []string
	// A := []int{11, 90, 8, 9, 7, 6}
	// fmt.Println("Array Before Insertion Sort Algorithm---->")
	// PrintArr1(&A)
	// Insertion1(&A)
	// fmt.Println("Array After Insertion Sort Algorithm---->")
	// PrintArr1(&A)
	loadFNames(&fnames)
	loadMNames(&mnames)
	//Read Last Name file
	file, err := os.Open("./isamg/data/last_names.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	file1, err := os.Open("./isamg/data/addresses.csv")
	if err != nil {
		fmt.Println("Error opening file1:", err)
		return
	}
	defer file1.Close()

	fileo, err := os.Create("./isamg/data/contacts.csv") //master file
	if err != nil {
		fmt.Println("Error opening file1:", err)
		return
	}
	defer fileo.Close()

	var ID string
	var idx int
	var gender int

	scanner := bufio.NewScanner(file)
	scanner1 := bufio.NewScanner(file1)
	var email string
	for scanner.Scan() { // Reading all last names
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		lastName := formatName(scanner.Text())

		var firstName string
		if gender == 0 {
			firstName = fnames[idx]
		} else {
			firstName = mnames[idx]
		}

		scanner1.Scan() // Read the next address record

		if err1 := scanner1.Err(); err1 != nil {
			fmt.Println("Error reading file:", err1)
			return
		}
		//TODO:  split csv string into an array then apply "scrub" to each field as needed.
		//       ex. removeDashes()
		//       email address may need to be lengthened as well as phone number.
		//       Ex. anne.straesser@scio-automation.com   45
		//       international phone numbers: +49(0)6233 6000-645   15 not counting parens, dashes
		//       Then concat array element into final csv AddressRec.
		//       remove ID sequence.
		//       write out contact version
		//       save writing out customer version for later.
		//
		//addrRec := removeDashes(scanner1.Text())
		addrRec := scanner1.Text()
		fields := strings.Split(addrRec, ",")
		if gender == 0 {
			email = firstN(firstName, 1) + "." + strings.ToLower(lastName) + "@earthlink.com"
		} else {
			email = firstN(firstName, 1) + "." + strings.ToLower(lastName) + "@scio-automation.com"
		}

		iPAddr := "123.24.15.100"
		msgDate := julian.GregorianToJulian(dateStmp) //strDate format CCYYMMDD
		message := "Test Message created in GenFile"

		// newidx2 = append(newidx2, 10)
		// fmt.Fprintf(idxFl2, "%s", newidx2) // write newrec to file
		fmt.Fprintf(fileo, "%s,%s,%s,%s,%s,%s,%s,%d,%s,%s,%s\n", ID, firstName, lastName, "", fields[6], email, iPAddr, msgDate, message, "", "")
		//fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%d,%s,%s,%s\n", ID, firstName, lastName, "", fields[6], email, iPAddr, msgDate, message, "", "")
		//ID++
		idx++
		//wrap the index back to 0
		if idx > 99 {
			idx = 0
			if gender == 0 {
				gender = 1
			} else {
				gender = 0
			}
		}
	}
}
