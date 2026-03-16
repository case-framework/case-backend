# Update mongodb driver

- The bson/primitive package has been merged into the bson package —> changed any instance of primitive.ObjectID to bson.ObjectId.

## All DB packages

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

- save scheduled-emails: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal

#### sms-templates

- save sms template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal

### participant user

#### user-attributes

- The UpdateOptions has been chnaged to UpdateOneOptions to configure UpdateOne operation.

#### users

- add user: MongoDB Go Driver v2 no longer allows constructing or modifying option structs directly, so update options must now be created through the new builder API (options.UpdateOne().SetUpsert(true)) instead of setting fields on UpdateOptions manually.
- update user in db: FindOneAndReplaceOptions can no longer be created or populated as a struct, so the v2 driver requires using the builder pattern (options.FindOneAndReplace().SetReturnDocument(options.After)) instead of setting option fields directly.
