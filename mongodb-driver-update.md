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
- removed unused return value for context.WithTimeout()

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

#### otps

- update the callback for mongo.WithSession to use a context.Context implementation, rather than the custom mongo.SessionContext

TEST: If you want to be extra safe, you can:
Deploy this change to a non‑production environment first.
Run a few test flows that:
exceed the maxOTPCount to ensure the “too many OTP requests” path still works,
run concurrent OTP creations to confirm only the allowed number of documents is written.

### study

#### participants

- configure FindOneAndReplaceOptions through options.FindOneAndReplace().Set... instead of filling the struct fields directly.

#### confidential responses

- configure replace options via the options.Replace() builder (for example, options.Replace().SetUpsert(true)) instead of instantiating a ReplaceOptions struct literal and passing its address.

#### study-rules

- &options.FindOneOptions{ Sort: sortByPublished } becomes options.FindOne().SetSort(sortByPublished).

#### reports

- GetUniqueReportKeysForStudy:`Distinct()` no longer returns `([]interface{}, error)`; it returns a single result type on which you call `.Decode(&target)` directly into a `[]string`, eliminating the manual type-assertion loop.

#### surveys

- GetSurveyKeysForStudy: `Distinct()` no longer returns `([]interface{}, error)`; it returns a single result type on which you call `.Decode(&target)` directly into a `[]string`, eliminating the manual type-assertion loop.
- GetCurrentSurveyVersion: create FindOneOptions using the options.FindOne() builder and setters (for example, options.FindOne().SetSort(sortByPublishedDesc)) instead of instantiating &options.FindOneOptions{} and mutating its fields.
- GetSurveyVersions: create FindOptions using the options.Find() builder and its setters (for example, options.Find().SetProjection(...).SetSort(...)) instead of instantiating &options.FindOptions{} and mutating its fields.
