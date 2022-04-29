module mail-go

go 1.16

require (
	github.com/zhangdapeng520/zdpgo_email/imap v0.0.0-20220119134953-dcd9ee65c8c7
	github.com/zhangdapeng520/zdpgo_email/message v0.15.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

//github.com/zhangdapeng520/zdpgo_email/imap v0.0.0-20220119134953-dcd9ee65c8c7 => ../pkg-replace/go-imap
//replace github.com/zhangdapeng520/zdpgo_email/message v0.15.0 => ../pkg-replace/go-message

require (
	golang.org/x/text v0.3.7
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)
