package types

type ContactPreferences struct {
	SubscribedToNewsletter        bool     `bson:"subscribedToNewsletter" json:"subscribedToNewsletter"`
	SendNewsletterTo              []string `bson:"sendNewsletterTo" json:"sendNewsletterTo"`
	SubscribedToWeekly            bool     `bson:"subscribedToWeekly" json:"subscribedToWeekly"`
	ReceiveWeeklyMessageDayOfWeek int32    `bson:"receiveWeeklyMessageDayOfWeek" json:"receiveWeeklyMessageDayOfWeek"`
}
