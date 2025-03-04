package isamg

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// At End Flags
var ErrAtEnd error = errors.New("End of Data")
var ErrNotFnd error = errors.New("Not Found")
var ErrCast error = errors.New("Invalid Type")
var ErrDUPKEY error = errors.New("KeyVal already on file")
var ErrRecChgd error = errors.New("Record changed since last read")

// Common constants
const DELFLAG string = "D"

// Lock Type operators
const ISLOCKNONE uint8 = 20
const ISLOCKMAN uint8 = 21
const ISLOCKEX uint8 = 22

// Compare type operators
const ISEQUAL uint8 = 1
const ISGREAT uint8 = 2
const ISGTEQ uint8 = 3

// Navigation type operators
const ISCURR uint8 = 11
const ISNEXT uint8 = 12
const ISPREV uint8 = 13
const ISFIRST uint8 = 14
const ISLAST uint8 = 15

// (Common types)
// IOState variables
type IOState struct {
	FD        *os.File //Pointer to opened master file
	Offset    int64    //Last read position of current file
	Idx       int      //Select index in the file group (0-master, 1-pri idx, 2-sec idx1 ... 10)
	MI        []byte   //entire index read into memory as a byte slice at this location
	MI_Offset int      //offset position in index (in-memory) where key was found
	Rrid      int64    //Relative Record number of the sought master file record
}

// Contact --------------------------------------
type Contact struct {
	ID        int    `fixed:"1,6,right,0"`
	FirstName string `fixed:"7,36"`
	LastName  string `fixed:"37,66"`
	MidInit   string `fixed:"67,67"`
	Phone     string `fixed:"68,87"`
	Email     string `fixed:"88,132"`
	IPAddr    string `fixed:"133,148"`
	MsgDate   int    `fixed:"149,155,right,0"`
	Message   string `fixed:"156,295"`
	RspDate   int    `fixed:"296,302,right,0"`
	DelFlag   string `fixed:"303,303"`
}
type Contact_id struct {
	ID      int    `fixed:"1,6,right,0"`
	Rrid    int    `fixed:"7,12,right,0"`
	DelFlag string `fixed:"13,13"`
}

type Contact_date struct {
	MsgDate int    `fixed:"1,7"`
	Name    string `fixed:"8,45"`
	Rrid    int    `fixed:"46,51,right,0"`
	RspDate int    `fixed:"52,58,right,0"`
	DelFlag string `fixed:"59,59"`
}
type Contact_email struct {
	Email   string `fixed:"1,45"`
	Rrid    int    `fixed:"46,51,right,0"`
	DelFlag string `fixed:"52,52"`
}

// For rewriting key only
type Contact_email_key struct {
	Email string `fixed:"1,45"`
}

// For rewriting key only
type Contact_id_key struct {
	ID int `fixed:"1,6,right,0"`
}

// For rewriting key only
type Contact_date_key struct {
	MsgDate int    `fixed:"1,7"`
	Name    string `fixed:"8,45"`
}

// Record struct Methods

func (cni *Contact) GetNameFmt() string {
	return strings.Trim(cni.LastName, " ") + ", " + strings.Trim(cni.FirstName, " ") + " " + cni.MidInit
}

// Returns Contact email
func (cn *Contact_email) GetCo_em_key(c *Contact) ([]byte, string) {
	return []byte(DELFLAG), c.Email
}

// Returns the Contact id
func (cn *Contact_id) GetCo_id_key(c *Contact) ([]byte, string) {
	return []byte(DELFLAG), fmt.Sprintf("%06d", c.ID)
}

// Returns the formatted Contact name
func (cn *Contact_date) GetCo_dt_key(c *Contact) ([]byte, string) {
	//return strconv.Itoa(c.MsgDate)
	return []byte(DELFLAG), fmt.Sprintf("%06d", c.MsgDate) + c.GetNameFmt()
}

// Return populated index record struct
func (cn *Contact_email) SetCo_em(c *Contact, rrid int) *Contact_email {
	cne := new(Contact_email)
	cne.Email = c.Email
	cne.Rrid = rrid
	return cne
}

// Return populated index record struct
func (cn *Contact_id) SetCo_id(c *Contact, rrid int) *Contact_id {
	cni := new(Contact_id)
	cni.ID = c.ID
	cni.Rrid = rrid
	return cni
}

// Return populated index record struct
func (cn *Contact_date) SetCo_dt(c *Contact, rrid int) *Contact_date {
	cnd := new(Contact_date)
	cnd.MsgDate = c.MsgDate
	cnd.Name = c.GetNameFmt()
	cnd.Rrid = rrid
	cnd.RspDate = c.RspDate
	return cnd
}

// Returns the new key portion of the Contact_date record if MsgDate or name changed on the master
func (cn *Contact_date) SetCo_dt_key(c *Contact, oldc *Contact) (*Contact_date_key, string) {
	cnk := new(Contact_date_key)
	cnk.MsgDate = c.MsgDate
	cnk.Name = c.GetNameFmt()
	//If no new data on key fields, signal no need for rewrite
	if cnk.Name == oldc.GetNameFmt() && cnk.MsgDate == oldc.MsgDate {
		return nil, ""
	}
	//keyVal is date and name concatenated
	return cnk, fmt.Sprintf("%06d", oldc.MsgDate) + oldc.GetNameFmt() //return new key(full col len), old key(trimmed)
}

// Returns the new key portion of the Contact_email record
func (cn *Contact_email) SetCo_em_key(c *Contact, oldc *Contact) (*Contact_email_key, string) {
	cnk := new(Contact_email_key)
	cnk.Email = c.Email
	//If no new data on key fields, signal no need for rewrite
	if cnk.Email == oldc.Email {
		return nil, ""
	}
	return cnk, oldc.Email //return new key(full col len), old key(trimmed)
}

// Returns the new key portion of the Contact_id record
func (cn *Contact_id) SetCo_id_key(c *Contact, oldc *Contact) (*Contact_id_key, string) {
	//changing ID value currently not supported
	return nil, ""

	// cnk := new(Contact_id_key)
	// cnk.ID = c.ID
	// //If no new data on key fields, signal no need for rewrite
	// if cnk.ID == oldc.ID {
	// 	return nil, ""
	// }
	// strID := fmt.Sprintf("%06d", oldc.ID) // Convert int to string and pad to 6 digits with leading zeros
	// return cnk, strID //return new key(full len), old key(full len)
}
