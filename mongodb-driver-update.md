# Update mongodb driver

- The bson/primitive package has been merged into the bson package —> changed any instance of primitive.ObjectID to bson.ObjectId.

## All DB packages

- context.Context parameter has been removed from mongo.Connect() because the deployment connector doesn’t accept a context, meaning that the context passed to mongo.Connect() in previous versions didn't serve a purpose.
- Simplfied DropOne and DropAll methods by removing the server response
- TESTING REQUIRED: Index Model: The old `IndexOptionsBuilder` type was removed and `IndexModel.Options.Name` is no longer accessible as a field. Required steps:
  - define variable for index names
  - capture index names returned by CreateMany
  - use the stored names when dropping indexes
  - maybe store indexNames as field of messagingDBService struct?!? (e.g. `emailTemplateIndexNames map[string][]`)
- removed unused return value for context.WithTimeout()

### Messaging

#### email-templates

- TESTING REQUIRED: save email template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (testing? Line 128/129)

#### scheduled-emails

- TESTING REQUIRED: save scheduled-emails: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal

#### sms-templates

- TESTING REQUIRED: save sms template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal

### participant user

#### user-attributes

- TESTING REQUIRED: The UpdateOptions has been chnaged to UpdateOneOptions to configure UpdateOne operation.

#### users

- add user: MongoDB Go Driver v2 no longer allows constructing or modifying option structs directly, so update options must now be created through the new builder API (options.UpdateOne().SetUpsert(true)) instead of setting fields on UpdateOptions manually. TESTING REQUIRED
- update user in db: FindOneAndReplaceOptions can no longer be created or populated as a struct, so the v2 driver requires using the builder pattern (options.FindOneAndReplace().SetReturnDocument(options.After)) instead of setting option fields directly. TESTING REQUIRED

#### otps

- update the callback for mongo.WithSession to use a context.Context implementation, rather than the custom mongo.SessionContext TESTING REQUIRED

TEST: If you want to be extra safe, you can:
Deploy this change to a non‑production environment first.
Run a few test flows that:
exceed the maxOTPCount to ensure the “too many OTP requests” path still works,
run concurrent OTP creations to confirm only the allowed number of documents is written.

### study

#### participants

- configure FindOneAndReplaceOptions through options.FindOneAndReplace().Set... instead of filling the struct fields directly. TESTING REQUIRED

#### confidential responses

- configure replace options via the options.Replace() builder (for example, options.Replace().SetUpsert(true)) instead of instantiating a ReplaceOptions struct literal and passing its address. TESTING REQUIRED

#### study-rules

- &options.FindOneOptions{ Sort: sortByPublished } becomes options.FindOne().SetSort(sortByPublished). TESTING REQUIRED

#### reports

- GetUniqueReportKeysForStudy:`Distinct()` no longer returns `([]interface{}, error)`; it returns a single result type on which you call `.Decode(&target)` directly into a `[]string`, eliminating the manual type-assertion loop.
TESTING REQUIRED:

#### surveys

- GetSurveyKeysForStudy: `Distinct()` no longer returns `([]interface{}, error)`; it returns a single result type on which you call `.Decode(&target)` directly into a `[]string`, eliminating the manual type-assertion loop.TESTING REQUIRED:
- GetCurrentSurveyVersion: create FindOneOptions using the options.FindOne() builder and setters (for example, options.FindOne().SetSort(sortByPublishedDesc)) instead of instantiating &options.FindOneOptions{} and mutating its fields. TESTING REQUIRED:
- GetSurveyVersions: create FindOptions using the options.Find() builder and its setters (for example, options.Find().SetProjection(...).SetSort(...)) instead of instantiating &options.FindOptions{} and mutating its fields. TESTING REQUIRED:
