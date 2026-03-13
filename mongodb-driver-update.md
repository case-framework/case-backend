# Update mongodb driver

- The bson/primitive package has been merged into the bson package —> changed any instance of primitive.ObjectID to bson.ObjectId.

## DB

- context.Context parameter has been removed from mongo.Connect() because the deployment connector doesn’t accept a context, meaning that the context passed to mongo.Connect() in previous versions didn't serve a purpose.
- Simplfied DropOne and DropAll methods by removing the server response
- Index Model: The old `IndexOptionsBuilder` type was removed and `IndexModel.Options.Name` is no longer accessible as a field. Required steps:
  - define variable for index names
  - capture index names returned by CreateMany
  - use the stored names when dropping indexes
  - maybe store indexNames as field of messagingDBService struct?!? (e.g. `emailTemplateIndexNames map[string][]`)

### Messaging

#### email-templates

- save email template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (testing? Line 128/129)

#### scheduled-emails

- save scheduled-emails: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (testing? Line 128/129)

#### sms-templates

- save sms template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (testing? Line 128/129)
