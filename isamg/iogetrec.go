// io file processing
package isamg

import (
	"fmt"
	"reflect"
	"isamg/fixedlen"
)

// Build File Record, returns appropriate []byte format record
func GetRec(rec interface{}, oldRec interface{}, idx int, rrid int, caller string) ([]byte, string, error) {
	var errST error

	// All input parameters are not used for each caller and keyVal is only returned for "R"
	// the ISREWRITE proc.
	//
	var keyVal string
	var rec1 interface{}
	var emptyrec []byte

	switch r := FileDf[idx].Rec.(type) {
	case *Contact: //maintain the master file record
		rec1 = rec
	case *Contact_email:
		switch caller {
		case "W": //ISWRITE ...
			rec1 = r.SetCo_em(rec.(*Contact), rrid) //populate index record
		case "R": //ISREWRITE
			rec1, keyVal = r.SetCo_em_key(rec.(*Contact), oldRec.(*Contact))
		case "D": //ISDELETE
			rec1, keyVal = r.GetCo_em_key(rec.(*Contact))
		}
	case *Contact_id:
		switch caller {
		case "W": //ISWRITE ...
			rec1 = r.SetCo_id(rec.(*Contact), rrid) //populate index record
		case "R": //ISREWRITE
			rec1, keyVal = r.SetCo_id_key(rec.(*Contact), oldRec.(*Contact))
		case "D": //ISDELETE
			rec1, keyVal = r.GetCo_id_key(rec.(*Contact))
		}
	case *Contact_date:
		switch caller {
		case "W": //ISWRITE ...
			rec1 = r.SetCo_dt(rec.(*Contact), rrid) //populate index record
		case "R": //ISREWRITE
			rec1, keyVal = r.SetCo_dt_key(rec.(*Contact), oldRec.(*Contact))
		case "D": //ISDELETE
			rec1, keyVal = r.GetCo_dt_key(rec.(*Contact))
		}
	//TODO: Check this assignment is it a copy from one instance to another?
	default:
		fmt.Println("Rec Type Unknown")
	}

	if rec1 == nil || reflect.ValueOf(rec1).IsNil() {
		return emptyrec, keyVal, errST
	}
	//In the case of ISDELETE & indexes (idx > 0), rec1 is already a []byte type, no need to marshal
	if caller == "D" && idx > 0 {
		return rec1.([]byte), keyVal, errST
	} else {
		return marshal(rec1), keyVal, errST
	}
}

func marshal(rec interface{}) []byte {
	var emptyrec []byte
	//TODO?:  Modify Marshal to add newline at end of byte slice if 2nd param is true
	newrec, err1 := fixedlen.Marshal(rec)
	if err1 == nil {
		return newrec
	} else {
		return emptyrec
	}
}
