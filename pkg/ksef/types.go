package ksef

import "time"

type SubjectType string

const (
	Subject1          SubjectType = "Subject1"
	Subject2          SubjectType = "Subject2"
	Subject3          SubjectType = "Subject3"
	SubjectAuthorized SubjectType = "SubjectAuthorized"
)

type InvoiceQuery struct {
	From        time.Time
	To          time.Time
	SubjectType SubjectType
}

type InvoicePage struct {
	PageSize   int32
	PageOffset int32
}
