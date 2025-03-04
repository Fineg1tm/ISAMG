package isamg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Insertion sort, in memory, on strings
func sort(A *[]string, keyl int) {
	for j := 1; j < len((*A)); j++ {
		i := j - 1
		rec := (*A)[j]
		key := (*A)[j][0:keyl]
		for i >= 0 && (*A)[i][0:keyl] > key {
			(*A)[i+1] = (*A)[i]
			i--
		}
		//(*A)[i+1] = key
		(*A)[i+1] = rec //swap the whole record, key is shorter for comparisons
	}
}

func isSorted(idx int) (bool, int64, error) {
	var err error
	var sorted bool = true

	// Open index files for read only
	// If computed record count is > than the hdr record count, then records have been added
	// since last sort.  Or if the sortFlag is valued 'S'  (due to a index key change), then a
	// sort is indicated.
	var totRecs int64
	var recCnt int
	var offset int
	//open index file denoted by idx
	fI, err := os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_RDONLY, 0644)
	if err == nil {
		fileinfo, err := fI.Stat()
		if err == nil {
			filesize := fileinfo.Size()
			totRecs = (filesize / (int64(FileDf[idx].LRecl) + 1))

			buffer := make([]byte, 6)
			offset = 6
			_, err = fI.ReadAt(buffer, int64(offset)) // read header recCnt
			if err != nil && err != io.EOF {
				fmt.Println(err)
				return true, 0, err
			}
			recCnt, err = strconv.Atoi(string(buffer))
			if err != nil {
				return true, 0, err
			}
			//---------------------------------
			buff1 := make([]byte, 1)
			offset = FileDf[idx].LRecl - 1
			_, err = fI.ReadAt(buff1, int64(offset)) // read sortFlag
			if err != nil && err != io.EOF {
				fmt.Println(err)
				return true, 0, err
			}
			sortFlag := string(buff1)

			//Examine indicators that a sort is needed.
			// 1.  Records appended since last sort
			// 2.  Index key values changed (ex. name change) and are now out of sequence
			if totRecs > int64(recCnt) || sortFlag == "S" {
				sorted = false
			}
		}
	}
	fI.Close()
	return sorted, totRecs, err
}

// Read a block of records into a string slice
func procBlock(blk *[]string, idx, recStart int) error {
	var err error
	maxRec := cap(*blk)

	// Open index file to sort
	// TODO: reference FileDf

	idxFile, inErr := os.Open(FileDf[idx].Path + FileDf[idx].Name) //open infile
	if inErr != nil {
		fmt.Println("Error opening file:", inErr)
		return inErr
	}
	defer idxFile.Close()
	// Seek to the desired record number.
	offset := int64(recStart) * (int64(FileDf[idx].LRecl) + 1)
	_, err = idxFile.Seek(offset, 0)
	if err != nil {
		return err
	}
	recCnt := 0
	var inrec string
	scanner := bufio.NewScanner(idxFile) //scan the contents of a file and print line by line
	for scanner.Scan() {
		if recCnt >= maxRec {
			break // Exit the loop when maxRec load count reached
		}
		inrec = scanner.Text()
		//Add record to slice
		*blk = append(*blk, inrec)
		recCnt += 1
	}
	return err
}

// Write block of sorted records to index file
func write(blk *[]string, idx int) error {
	var errWR error

	//open output file
	fD, errOp := os.Create(FileDf[idx].Path + FileDf[idx].Name + ".tmp")
	if errOp != nil {
		return errOp
	}

	var r int
	for r = 0; r < len(*blk); r++ {
		rec := []byte((*blk)[r])
		rec = append(rec, 10)
		_, errWR = fD.Write(rec) //sort out writes
		if errWR != nil {
			fmt.Println(errWR)
			//log.Fatal(errWR)
		}
	}
	//Update the recCnt and the sort flag on the header record
	count := []byte(fmt.Sprintf("%06d", r-1))
	sortFlag := []byte{32}

	_, errWR = fD.WriteAt(count, 6) //Update hdr rec count
	if errWR != nil {
		fmt.Println(errWR)
	}
	offset := int64(FileDf[idx].LRecl - 1)  //offset to last byte of the logical record
	_, errWR = fD.WriteAt(sortFlag, offset) //Clear hdr sortFlag
	if errWR != nil {
		fmt.Println(errWR)
	}
	fD.Close() //close immediately, file will be renamed next

	// Move old file to backup directory
	fromDir := FileDf[idx].Path + FileDf[idx].Name
	toDir := FileDf[idx].Path + "bkp/" + FileDf[idx].Name
	errMV := os.Rename(fromDir, toDir)
	if errMV == nil {
		// Remove .tmp extension on new sorted file to complete replacement
		fromFile := FileDf[idx].Path + FileDf[idx].Name + ".tmp"
		toFile := FileDf[idx].Path + FileDf[idx].Name
		errRN := os.Rename(fromFile, toFile)
		if errRN != nil {
			fmt.Println(errRN)
		}
	} else {
		fmt.Println(errMV)
	}

	return errWR
}

func ISSORT(fileGrp string) {
	//Load the file group

	getFDefs(fileGrp)

	//Loop through the list, check to see which indexes need to be sorted
	var fileCnt int = len(FileDf)
	for idx := 1; idx < fileCnt; idx++ {
		sorted, totRecs, sErr := isSorted(idx)
		if sErr != nil {
			fmt.Println("A problem processing index header record occured", idx, sErr.Error())
			return
		}
		// Sort Indexes that are out of sort
		if !sorted {
			keyl := FileDf[idx].Keyl
			block1 := make([]string, 0, totRecs)

			procBlock(&block1, idx, 0)
			sort(&block1, keyl)
			write(&block1, idx)
		}
	}
}
