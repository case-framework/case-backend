package types

type ContactPreferences struct {
	SubscribeToNewsletter         bool     `bson:"subscribeToNewsletter" json:"subscribeToNewsletter"`
	SendNewsletterTo              []string `bson:"sendNewsletterTo" json:"sendNewsletterTo"`
	SubscribedToWeekly            bool     `bson:"subscribedToWeekly" json:"subscribedToWeekly"`
	ReceiveWeeklyMessageDayOfWeek int32    `bson:"receiveWeeklyMessageDayOfWeek" json:"receiveWeeklyMessageDayOfWeek"`
}
