// io functions of the isamg file system.
package isamg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"isamg/fixedlen"
	"isamg/seq"
)

var FileDfs = make(FDefs) //Table for definitions of ALL files by name+index
var FileDf FileDefs       //Array for definitions of SELECTED file group by numeric index
var IoS *IOState          //IO State: current working files, byte slices and pointers

func init() {
	//Using init to populate File Definition map.
	
	basedir := os.Getenv("GOPATH")
	subDir := "isamg_data"
	path := filepath.Join(basedir, subDir, "-")
	path = path[:len(path)-1] //no dash, just the final slash

	FileDfs["contact0"] = new(FileDef)
	FileDfs["contact0"].Name = "contact.dat"
	FileDfs["contact0"].Path = path
	FileDfs["contact0"].Rec = new(Contact)
	FileDfs["contact0"].LRecl = 303
	FileDfs["contact0"].Keyl = 6
	FileDfs["contact0"].Rridl = 0

	FileDfs["contact1"] = new(FileDef)
	FileDfs["contact1"].Name = "contact_email.idx"
	FileDfs["contact1"].Path = path
	FileDfs["contact1"].Rec = new(Contact_email)
	FileDfs["contact1"].LRecl = 52
	FileDfs["contact1"].Keyl = 45
	FileDfs["contact1"].Keyo = 88 //column where key begins on master file
	FileDfs["contact1"].Keyu = true
	FileDfs["contact1"].Rridl = 6

	FileDfs["contact2"] = new(FileDef)
	FileDfs["contact2"].Name = "contact_id.idx"
	FileDfs["contact2"].Path = path
	FileDfs["contact2"].Rec = new(Contact_id)
	FileDfs["contact2"].LRecl = 13
	FileDfs["contact2"].Keyl = 6
	FileDfs["contact2"].Rridl = 6

	FileDfs["contact3"] = new(FileDef)
	FileDfs["contact3"].Name = "contact_date.idx"
	FileDfs["contact3"].Path = path
	FileDfs["contact3"].Rec = new(Contact_date)
	FileDfs["contact3"].LRecl = 59
	FileDfs["contact3"].Keyl = 45
	FileDfs["contact3"].Rridl = 6

}

// IoS - IO State initialized.  TODO / Test that instances are independent across threads
func InitIoS() *IOState {
	IoS = new(IOState)
	FileDf = make(FileDefs, 0, 11) //Array for definitions of SELECTED file group by numeric index
	return IoS
}
func ResetIoS() {
	IoS = new(IOState)
}

// Open an index (primary by default)
func ISOPEN0(fileGrp string, openType int, lockType uint8) error {

	//Load the targeted files' definitions
	getFDefs(fileGrp)

	var mI []byte
	var readLen int
	seq.Initialize(fileGrp, FileDf[0].Path)
	fD, err := os.OpenFile(FileDf[0].Path+FileDf[0].Name, openType, 0644) //master file
	if err == nil {
		if openType != os.O_APPEND|os.O_WRONLY {
			fI, err := os.OpenFile(FileDf[IoS.Idx].Path+FileDf[IoS.Idx].Name, openType, 0644) //index file
			if err == nil {
				fileinfo, err := fI.Stat()
				if err == nil {
					filesize := fileinfo.Size()
					mI = make([]byte, filesize)
					readLen, err = fI.Read(mI)
					if readLen <= 0 || err != nil {
						fmt.Println(err, "mI ReadLen: ", readLen)
						return err
					}
				}
			}
		}
	}
	IoS.FD = fD
	IoS.MI = mI
	return err
}

// Open an index (selected in application)
func ISOPEN(fileGrp string, openType int, lockType uint8) error {

	//Load the targeted files' definitions
	getFDefs(fileGrp)

	var fD *os.File
	var mI []byte
	var readLen int
	seq.Initialize(fileGrp, FileDf[IoS.Idx].Path)

	//open index file and read entire file into memory
	fI, err := os.OpenFile(FileDf[IoS.Idx].Path+FileDf[IoS.Idx].Name, os.O_RDONLY, 0644)
	if err == nil {
		fileinfo, err := fI.Stat()
		if err == nil {
			filesize := fileinfo.Size()
			mI = make([]byte, filesize)
			readLen, err = fI.Read(mI)
			if readLen <= 0 || err != nil {
				fmt.Println(err, "mI ReadLen: ", readLen)
				return err
			}
			IoS.MI = mI
		}
	}
	if openType == os.O_RDONLY {
		//together with driving index, open master file to read data in a particular order
		//Otherwise, for Singleton Adds and Updates, files in a file group will be opened/Read JIT
		//for write and rewrite.
		fD, err = os.OpenFile(FileDf[0].Path+FileDf[0].Name, openType, 0644) //master file
		if err != nil {
			fmt.Println(err, "unable to open: ", fileGrp)
			return err
		}
		IoS.FD = fD
	}
	return err
}

func ISSTART(keyVal string, compareType uint8) error {

	//var mI_Offset int
	offset := LinearSrch(IoS.MI, []byte(keyVal), FileDf[IoS.Idx].LRecl+1) //seach for first occurrence of keyVal
	if offset == -1 {
		IoS.Rrid = -1
		return ErrNotFnd
	}
	//Save the Index offset
	IoS.MI_Offset = offset
	// Advance offset to rrid position and return as an integer
	offset += FileDf[IoS.Idx].Keyl
	rrid, err := strconv.ParseInt(string(IoS.MI[offset:offset+FileDf[IoS.Idx].Rridl]), 10, 64)
	if err != nil {
		fmt.Println(err)
		return err
	}
	IoS.Rrid = rrid
	return err
}
func ISREAD(keyVal string, v interface{}, navType uint8) error {
	sep := []byte(keyVal) //search used to limit ISNEXT reads
	n := len(sep)

	if navType == ISNEXT {
		// if Offset is valid and not at end of index buffer and not beyond search value
		if IoS.MI_Offset >= 0 && IoS.MI_Offset < len(IoS.MI) &&
			Equal(IoS.MI[IoS.MI_Offset:IoS.MI_Offset+n], sep) {

			tmpOffset := IoS.MI_Offset + FileDf[IoS.Idx].Keyl //point to rrid col
			next_rrid, err := strconv.ParseInt(string(IoS.MI[tmpOffset:tmpOffset+FileDf[IoS.Idx].Rridl]), 10, 64)
			if err != nil {
				fmt.Println(err)
				return err
			}
			IoS.Rrid = next_rrid
			IoS.MI_Offset += (FileDf[IoS.Idx].LRecl + 1) // ISNEXT: Advance IoS.MI_Offset to next idx record
		} else {
			return ErrAtEnd // End of Data
		}
	}
	buffer := make([]byte, FileDf[0].LRecl)           // buffer for master file record read
	offset := (int64(FileDf[0].LRecl) + 1) * IoS.Rrid // Add 1 for the NewLine char
	readLen, readErr := IoS.FD.ReadAt(buffer, offset) // read record at offset position
	fmt.Println("readLen: ", readLen)
	if readErr != nil && readErr != io.EOF {
		fmt.Println(readErr)
		return readErr
	}
	err := fixedlen.Unmarshal(buffer, v)

	if err != nil {
		fmt.Println(err)
		return err
	}
	//Store the current record's offset for subsequent use.  Ex. REWRITE
	IoS.Offset = offset
	return err
}

// Adds records to the opened master file and automatically to its linked indexes.
func ISWRITE0(rec interface{}) error {
	var errWR error

	// Write to the master file
	// Save the indexed fields on the masterfile to build indexes below
	//
	var rrid int
	var rec1 interface{}
	var fileCnt int = len(FileDf)
	for idx := 0; idx < fileCnt; idx++ { //start with index 1 write 0 master last
		switch v1 := FileDf[idx].Rec.(type) {
		case *Contact:
			rec1 = rec //populate master record
		case *Contact_email:
			v1.SetCo_em(rec.(*Contact), rrid) //populate index record
			rec1 = v1
		case *Contact_id:
			v1.SetCo_id(rec.(*Contact), rrid) //populate index record
			rec1 = v1
		case *Contact_date:
			v1.SetCo_dt(rec.(*Contact), rrid) //populate index record
			rec1 = v1
		//TODO: Check this assignment is it a copy from one instance to another?
		default:
			fmt.Println("Type Unknown")
		}

		fD, errOp := os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_APPEND|os.O_WRONLY, 0644)

		if errOp != nil {
			fmt.Println(errOp)
			//log.Fatal(err1)
			return errOp
		}
		// get last rrid on master file
		if idx == 0 {
			fileinfo, _ := fD.Stat()
			rrid = (int(fileinfo.Size()) / (FileDf[idx].LRecl + 1))
		}

		//TODO:  Modify Marshal to add newline if true is 2nd param
		newrec, err1 := fixedlen.Marshal(rec1)
		if err1 != nil {
			fmt.Println(err1)
			//log.Fatal(err1)
		}
		newrec = append(newrec, 10)
		_, errWR = fD.Write(newrec) //writing index records and master in a loop
		if errWR != nil {
			fmt.Println(errWR)
			//log.Fatal(errWR)
		}
		fD.Close()
	}
	return nil
}

// Adds records to the opened master file and automatically to its linked indexes.
func ISWRITE(rec interface{}) error {
	var errWR error

	// Write to the master file
	// Save the indexed fields on the masterfile to build indexes below
	//
	var rrid int
	var oldRec interface{}
	var newrec []byte
	var fileCnt int = len(FileDf)

	for idx := 0; idx < fileCnt; idx++ { //start with index 1 write 0 master last

		newrec, _, _ = GetRec(rec, oldRec, idx, rrid, "W")

		fD, errOp := os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_APPEND|os.O_WRONLY, 0644)
		if errOp != nil {
			fmt.Println(errOp)
			//log.Fatal(err1)
			return errOp
		}
		// get last rrid on master file
		if idx == 0 {
			fileinfo, _ := fD.Stat()
			rrid = (int(fileinfo.Size()) / (FileDf[idx].LRecl + 1))
		}

		//TODO:  Modify Marshal to add newline if true is 2nd param
		// newrec, err1 := fixedlen.Marshal(rec1)
		// if err1 != nil {
		// 	fmt.Println(err1)
		// 	//log.Fatal(err1)
		// }
		newrec = append(newrec, 10)
		_, errWR = fD.Write(newrec) //writing index records and master in a loop
		if errWR != nil {
			fmt.Println(errWR)
			//log.Fatal(errWR)
		}
		fD.Close()
	}
	return nil
}

// Batch - Adds records to already open files in FileGrp.  rrid incremented on successful write to master
func ISWRITEB(rec interface{}, rrid int) error {
	var errWR error

	var oldRec interface{}
	var newrec []byte
	var fileCnt int = len(FileDf)

	for idx := 0; idx < fileCnt; idx++ {

		newrec, _, _ = GetRec(rec, oldRec, idx, rrid, "W")

		newrec = append(newrec, 10)
		_, errWR = FileDf[idx].FD.Write(newrec) //writing index records and master in a loop
		if errWR != nil {
			fmt.Println(errWR)
			//log.Fatal(errWR)
		}
	}
	return nil
}

// Batch Update Index Header records total recCnts
func ISWRITEBT(recCnt int) error {
	var errWR error
	var fileCnt int = len(FileDf)

	count := []byte(fmt.Sprintf("%06d", recCnt))
	//Update index header records
	for idx := 1; idx < fileCnt; idx++ {
		_, errWR = FileDf[idx].FD.WriteAt(count, 6)
		if errWR != nil {
			fmt.Println(errWR)
			//log.Fatal(errWR)
		}
	}
	return nil
}

// Updates records on the opened master file and automatically on its linked indexes.
func ISREWRITE0(rec interface{}, oldRec interface{}) error {
	var errRWR error

	//TODO:  This assertion process for the record structures should be made global to the app
	//       callable from where needed.  ISWRITE, ISREWRITE, ISDELETE ...
	//       Same applies to the process of identifing key column data changes

	var rec1 interface{}
	var fileCnt int = len(FileDf)
	var keyVal string

	//loop to write data to master and assoc. indexes
	for idx := 0; idx < fileCnt; idx++ {
		switch v1 := FileDf[idx].Rec.(type) {
		case *Contact:
			rec1 = rec
		case *Contact_id:
			//  changing numeric ID disallowed
			// 	v1.SetCo_id(rec.(*Contact), rrid)
			// 	rec1 = v1
			rec1 = nil
			keyVal = ""
		case *Contact_email:
			rec1, keyVal = v1.SetCo_em_key(rec.(*Contact), oldRec.(*Contact))
		case *Contact_date:
			rec1, keyVal = v1.SetCo_dt_key(rec.(*Contact), oldRec.(*Contact))
		default:
		}
		//TODO:  The open below should be for idx files only.  For idx == 0 follow the 3 steps in
		//       test.go/updRec ISOPEN, ISSTART, READ to get a populated struct.
		//       Then compare it with original old. If they are not the same then exit with an error (data)
		//       interim update, start over.  If they are the same, procede with the update.  This is where
		//       locking for master and index updates should occur
		if rec1 == nil || reflect.ValueOf(rec1).IsNil() {
			continue // Skip this iteration, index not updated
		}

		var errIO error
		IoS.FD, errIO = os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_RDWR, 0644) //master or index file
		if errIO != nil {
			fmt.Println(errIO)
			//log.Fatal(errIO)
			return errIO
		}
		// check to ensure record hasn't changed in the interim.  If it hasn't then set LOCK
		// TODO set a LOCK
		if idx == 0 {
			notOK, _ := chkRecStatus(oldRec)
			if notOK {
				return ErrRecChgd //Abort update with a message
			}
		}
		if idx > 0 { //search index for key to get offset of record to rewrite
			IoS.Offset, errIO = searchIdx(IoS.FD, keyVal, idx)
			if errIO != nil || IoS.Offset == -1 {
				fmt.Println("Error: ", errIO, " Offset: ", IoS.Offset, " Idx: ", idx)
				//log.Fatal("Error: ", errIO, " Offset: ", IoS.Offset, " Idx: ", idx)
				continue
			}
		}

		newrec, err1 := fixedlen.Marshal(rec1)
		if err1 != nil {
			fmt.Println(err1)
			//log.Fatal(err1)
		}
		//newrec = append(newrec, 10)                  //using WriteAt below, not a whole record write
		_, errRWR = IoS.FD.WriteAt(newrec, IoS.Offset) //writing master and index records in a loop
		if errRWR != nil {
			fmt.Println(errRWR)
			//log.Fatal(errWR)
		}
		//---------------------------------
		//Set the sortFlag value "S" on the index hdr recs
		if idx > 0 {
			sortFlag := "S"
			buff1 := []byte(sortFlag)
			offset := int64(FileDf[idx].LRecl - 1)
			_, errRWR = IoS.FD.WriteAt(buff1, offset) //update sortFlag on hdr rec
			if errRWR != nil {
				fmt.Println(errRWR)
				//log.Fatal(errWR)
			}
		}
		IoS.FD.Close()
	}
	return nil
}

// Updates records on the opened master file and automatically on its linked indexes.
func ISREWRITE(rec interface{}, oldRec interface{}) error {
	var errRWR error

	//TODO:  This assertion process for the record structures should be made global to the app
	//       callable from where needed.  ISWRITE, ISREWRITE, ISDELETE ...
	//       Same applies to the process of identifing key column data changes

	var newrec []byte
	var fileCnt int = len(FileDf)
	var keyVal string

	//loop to write data to master and assoc. indexes
	for idx := 0; idx < fileCnt; idx++ {

		newrec, keyVal, _ = GetRec(rec, oldRec, idx, 0, "R")

		if len(newrec) == 0 {
			continue // Skip this iteration, index not updated
		}

		var errIO error
		IoS.FD, errIO = os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_RDWR, 0644) //master or index file
		if errIO != nil {
			fmt.Println(errIO)
			//log.Fatal(errIO)
			return errIO
		}
		// check to ensure record hasn't changed in the interim.  If it hasn't then set LOCK
		// TODO set a LOCK
		if idx == 0 {
			notOK, _ := chkRecStatus(oldRec)
			if notOK {
				return ErrRecChgd //Abort update with a message
			}
		}
		if idx > 0 { //search index for key to get offset of record to rewrite
			IoS.Offset, errIO = searchIdx(IoS.FD, keyVal, idx)
			if errIO != nil || IoS.Offset == -1 {
				fmt.Println("Error: ", errIO, " Offset: ", IoS.Offset, " Idx: ", idx)
				//log.Fatal("Error: ", errIO, " Offset: ", IoS.Offset, " Idx: ", idx)
				continue
			}
		}

		//newrec = append(newrec, 10)                  //using WriteAt below, not a whole record write
		_, errRWR = IoS.FD.WriteAt(newrec, IoS.Offset) //writing master and index records in a loop
		if errRWR != nil {
			fmt.Println(errRWR)
			//log.Fatal(errWR)
		}
		//---------------------------------
		//Set the sortFlag value "S" on the index hdr recs
		if idx > 0 {
			sortFlag := "S"
			buff1 := []byte(sortFlag)
			offset := int64(FileDf[idx].LRecl - 1)
			_, errRWR = IoS.FD.WriteAt(buff1, offset) //update sortFlag on hdr rec
			if errRWR != nil {
				fmt.Println(errRWR)
				//log.Fatal(errWR)
			}
		}
		IoS.FD.Close()
	}
	return nil
}

// Flags the master file record and all linked index file records for deletion.
func ISDELETE0(rec interface{}, oldRec interface{}) error {
	var errRWR error

	//TODO:  This assertion process for the record structures should be made global to the app
	//       callable from where needed.  ISWRITE, ISREWRITE, ISDELETE ...
	//       Same applies to the process of identifing key column data changes

	var offset int64
	var rec1 interface{}
	var fileCnt int = len(FileDf)
	var keyVal string

	//loop to write data to master and assoc. indexes
	for idx := 0; idx < fileCnt; idx++ {
		switch v1 := FileDf[idx].Rec.(type) {
		case *Contact:
			rec1 = rec //use master record struct passed from the client
		case *Contact_email:
			rec1, keyVal = v1.GetCo_em_key(rec.(*Contact))
		case *Contact_id:
			rec1, keyVal = v1.GetCo_id_key(rec.(*Contact))
		case *Contact_date:
			rec1, keyVal = v1.GetCo_dt_key(rec.(*Contact))
		default:
			fmt.Println(keyVal)
		}

		//TODO:  The open below should be for idx files only.  For idx == 0 follow the 3 steps in
		//       test.go/updRec ISOPEN, ISSTART, READ to get a populated struct.
		//       Then compare it with original old. If they are not the same then exit with an error (data)
		//       interim update, start over.  If they are the same, procede with the update.  This is where
		//       locking for master and index updates should occur
		//       ISDELETE and ISREWRITE are nearly the same procedure, except ISDELETE updates all indexes
		//       but look into merging and handling difference

		if rec1 != nil {
			var errIO error
			IoS.FD, errIO = os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_RDWR, 0644) //index file
			if errIO != nil {
				fmt.Println(errIO)
				//log.Fatal(errIO)
				return errIO
			}
			if idx > 0 { //search index for key to get offset of record to rewrite
				keyVal, _ := rec1.(string)
				offset, errIO = searchIdx(IoS.FD, keyVal, idx)
				//change the offset from first position to last (where the DelFlag is)
				IoS.Offset = offset + (int64(FileDf[idx].LRecl) - 1)
				if errIO != nil {
					fmt.Println(errIO)
					//log.Fatal(errIO)
					return errIO
				}
			}
		}
		//}
		if rec1 == nil {
			continue // Skip this iteration, index not updated
		}

		// For index record "DELETE", just the DelFlag is written to the record
		newrec := []byte(DELFLAG) //index file update (DelFlag changed)

		// Master file is written first (idx == 0).  Entire record with DelFlag is rewritten
		// NOTE: IoS.Offset for master is coming from back at first ISREAD, but it will be refreshed
		// when ISREAD for Update with lock is added above.

		if idx == 0 {
			var err1 error
			newrec, err1 = fixedlen.Marshal(rec1) //master file update (whole record)
			if err1 != nil {
				fmt.Println(err1)
				//log.Fatal(err1)
			}
		}

		_, errRWR = IoS.FD.WriteAt(newrec, IoS.Offset) //writing master and index records in a loop
		if errRWR != nil {
			fmt.Println(errRWR)
			//log.Fatal(errWR)
		}
		IoS.FD.Close()
	}
	return nil
}

// Flags the master file record and all linked index file records for deletion.
func ISDELETE(rec interface{}, oldRec interface{}) error {
	var errRWR error

	//TODO:  This assertion process for the record structures should be made global to the app
	//       callable from where needed.  ISWRITE, ISREWRITE, ISDELETE ...
	//       Same applies to the process of identifing key column data changes

	var offset int64
	var newrec []byte
	var keyVal string
	var fileCnt int = len(FileDf)

	//loop to write data to master and assoc. indexes
	for idx := 0; idx < fileCnt; idx++ {

		newrec, keyVal, _ = GetRec(rec, oldRec, idx, 0, "D")

		//TODO:  The open below should be for idx files only.  For idx == 0 follow the 3 steps in
		//       test.go/updRec ISOPEN, ISSTART, READ to get a populated struct.
		//       Then compare it with original old. If they are not the same then exit with an error (data)
		//       interim update, start over.  If they are the same, procede with the update.  This is where
		//       locking for master and index updates should occur
		//       ISDELETE and ISREWRITE are nearly the same procedure, except ISDELETE updates all indexes
		//       but look into merging and handling difference

		if len(newrec) > 0 {
			var errIO error
			IoS.FD, errIO = os.OpenFile(FileDf[idx].Path+FileDf[idx].Name, os.O_RDWR, 0644) //index file
			if errIO != nil {
				fmt.Println(errIO)
				//log.Fatal(errIO)
				return errIO
			}
			if idx > 0 { //search index for key to get offset of record to rewrite
				offset, errIO = searchIdx(IoS.FD, keyVal, idx)
				//change the offset from first position to last (where the DelFlag is)
				IoS.Offset = offset + (int64(FileDf[idx].LRecl) - 1)
				if errIO != nil {
					fmt.Println(errIO)
					//log.Fatal(errIO)
					return errIO
				}
			}
		}

		if len(newrec) < 1 {
			continue // Skip this iteration, index not updated
		}

		// Master file is written first (idx == 0).  Entire record with DelFlag is rewritten
		// NOTE: IoS.Offset for master is coming from back at first ISREAD, but it will be refreshed
		// when ISREAD for Update with lock is added above.

		_, errRWR = IoS.FD.WriteAt(newrec, IoS.Offset) //writing master and index records in a loop
		if errRWR != nil {
			fmt.Println(errRWR)
			//log.Fatal(errWR)
		}
		IoS.FD.Close()
	}
	return nil
}

// Set FileDf and IoS structures to nil for garbage collection
// TODO:  Pass fD, fI for close?
func ISCLOSE() {
	FileDf = nil
	IoS = nil
}

// Copy selected current file group to array for numeric index reference.  In addition to the fileGrp name,
// an alternate path to the files may be supplied.
func getFDefs(fileVals ...string) error {
	var path string
	fileGrp := fileVals[0]
	if len(fileVals) > 1 {
		path = fileVals[1]
	}
	var max int = 11 //TODO:  Remove the max variable once len() value is verified
	for i := 0; i < max; i++ {

		fileNme := fileGrp + strconv.Itoa(i)
		// Check if the key exists first to prevent panic
		if dfs, ok := FileDfs[fileNme]; ok {
			if len(path) > 0 {
				dfs.Path = path //override the default Path
			}
			FileDf = append(FileDf, *dfs)
		} else {
			break
		}
	}
	return nil
}

// Searches for the key value on file to find the record's offest from start of file
func searchIdx(fI *os.File, keyVal string, idx int) (int64, error) {
	var offset int
	var mI []byte
	//
	//Read index file into memory []byte
	fileinfo, err := fI.Stat()
	if err == nil {
		filesize := fileinfo.Size()
		mI = make([]byte, filesize)
		readLen, err := fI.Read(mI)
		if readLen <= 0 || err != nil {
			fmt.Println(err, "mI ReadLen: ", readLen)
			return -1, err
		}
	}
	//search index for first occurencee of key
	offset = LinearSrch(mI, []byte(keyVal), FileDf[idx].LRecl+1) //seach for first occurrence of keyVal

	// _, ok = FileDf[idx].Rec.(*Customer_name)
	// if ok {
	// 	offset = 630
	// } else {
	// 	offset = 360
	// }
	return int64(offset), nil
}

func chkRecStatus(oldRec interface{}) (bool, error) {
	// Client pulled the record earlier to start an update, then disconnected.
	// If the master file has not changed since reconnecting, than the update can proceed.
	var recChanged bool = true

	var keyVal string = "" //this value is used to stop ISNEXT loop.  Not used since ISEQUAL is specified on the read below.
	var currRec = new(Contact)

	//ISREAD is using IoS.Rrid preserved from original client's read/start
	rc := ISREAD(keyVal, currRec, ISEQUAL)
	if rc != nil {
		fmt.Println(rc)
		return true, rc
	}
	rec, _ := oldRec.(*Contact) //assert type
	if *rec == *currRec {
		recChanged = false
	}
	// switch rec := oldRec.(type) {
	// case *Contact:
	// 	if *rec == *currRec {
	// 		recChanged = false
	// 	}
	// }
	return recChanged, nil
}
