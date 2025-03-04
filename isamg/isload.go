// isload function populates and a master file and its indexes from a csv data file.
package isamg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"isamg/julian"
	"isamg/seq"
	"time"
)

// ISLOAD will write/replace data in the master file and index files of a File group.
func ISLOAD(fileGrp string, openType int, lockType uint8, path string, infile string, maxRec int) error {

	now := time.Now()
	runDate := now.Format("2006-01-02")
	DateStmp := strings.ReplaceAll(runDate, "-", "")
	//Load the targeted files' definitions
	getFDefs(fileGrp, path)

	var err error
	seq.Initialize(fileGrp, path)
	//***********************************************************************************
	// Load records from a csv file and populate the master file and its associated
	// indexes in a batch mode.  Start by opening all the files to be loaded ahead of time.
	//***********************************************************************************
	// Open output files
	var fileCnt int = len(FileDf)

	for idx := 0; idx < fileCnt; idx++ { //open files in this file group
		fD, errOp := os.Create(FileDf[idx].Path + FileDf[idx].Name)
		if errOp != nil {
			return errOp
		}
		FileDf[idx].FD = fD //save Open file Pointers
		defer FileDf[idx].FD.Close()

		//Write Header record 000000000000 with last 6 digits are the record count.
		hdrStat := []byte{48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48}
		if idx > 0 {
			hdrRec := make([]byte, FileDf[idx].LRecl)
			for i := 0; i < len(hdrRec); i++ {
				if i < len(hdrStat) {
					hdrRec[i] = hdrStat[i]
				} else {
					hdrRec[i] = 32
				}
			}
			hdrRec = append(hdrRec, 10)
			_, errWR := FileDf[idx].FD.Write(hdrRec) //writing index records and master in a loop
			if errWR != nil {
				fmt.Println(errWR)
			}
		}
		FileDf[idx].FD.Sync() //flush write to disk
	}
	// Open input data file
	datFile, inErr := os.Open(path + infile) //open infile
	if inErr != nil {
		fmt.Println("Error opening file:", inErr)
		return inErr
	}
	defer datFile.Close()

	recCnt := 0
	scanner := bufio.NewScanner(datFile) //scan the contents of a file and print line by line
	for scanner.Scan() {
		if recCnt >= maxRec {
			break // Exit the loop when maxRec load count reached
		}
		inrec := scanner.Text()
		field := strings.Split(inrec, ",")
		//---------------------
		var fRec = new(Contact)
		fRec.ID = seq.Next(fileGrp, path)
		fRec.FirstName = field[1]
		fRec.LastName = field[2]
		fRec.MidInit = field[3]
		fRec.Phone = field[4]
		fRec.Email = field[5]
		fRec.IPAddr = field[6]
		//field[7]
		fRec.MsgDate = julian.GregorianToJulian(DateStmp) //strDate format CCYYMMDD
		fRec.Message = field[8]
		//field[9]
		//fRec.RspDate = ""
		//field[10]
		//fRec.DelFlag = ""

		rc := ISWRITEB(fRec, recCnt) //ISWRITEB - batch write (files pre-opened and locked)
		if rc != nil {
			fmt.Println(rc)
			return rc
		}
		//fmt.Printf("CID: %d Last: %s First: %s Mid: %s Phone: %s Email: %s IP: %s MsgDate: %d Msg: %s \n", fRec.ID, fRec.LastName, fRec.FirstName, fRec.MidInit, fRec.Phone, fRec.Email, fRec.IPAddr, fRec.MsgDate, fRec.Message)

		// idxGen, castErr := strconv.ParseInt((fields[3]), 10, 16) //base 10, 16bit integer
		// if castErr != nil {
		// 	idxGen = 3
		// }

		//fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6], fields[7], fields[8], fields[9], fields[10])
		//fmt.Println(line)
		recCnt += 1
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from file:", err) //print error if scanning is not done properly
	}
	//Assumption at this point, that all indexes had the same number of writes that succeeded
	//Update the header record on the indexes
	rc := ISWRITEBT(recCnt) //ISWRITEB - batch write (files pre-opened and locked)
	if rc != nil {
		fmt.Println(rc)
		return rc
	}
	return err
}
