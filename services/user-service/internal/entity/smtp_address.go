package entity

type Email string

type SMTPType string

const (
	Primary   SMTPType = "Primary"
	Secondary SMTPType = "Secondary"
)

type SMTPAddress struct {
	Address  Email
	Identity string
	Type     SMTPType
}
