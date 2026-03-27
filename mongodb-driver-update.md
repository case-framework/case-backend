# Update mongodb driver

- The bson/primitive package has been merged into the bson package —> changed any instance of primitive.ObjectID to bson.ObjectId.

## All DB packages

- context.Context parameter has been removed from mongo.Connect() because the deployment connector doesn’t accept a context, meaning that the context passed to mongo.Connect() in previous versions didn't serve a purpose.
- Simplfied DropOne and DropAll methods by removing the server response
- removed unused return value for context.WithTimeout()
- Index Model: The old `IndexOptionsBuilder` type was removed and `IndexModel.Options.Name` is no longer accessible as a field. Required steps:
  - define static default index name constants/lists per collection
  - reuse those names for both index creation and `DropOne` in `drop defaults`

### Messaging

#### email-templates

- save email template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (testing? Line 128/129)

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

## Manual Test Protocol (Index Migration)

Date: 24.03.2026

Goal:

- Preserve MongoDB driver v1 behavior for default index drop/recreate flows after migration to driver v2.
- Verify that custom manually added indexes are not affected by default-index drop.

Executed commands:

```bash
CONFIG_FILE_PATH=test/jobs/dbm-01-before.yaml go run ./jobs/db-migration/*.go
CONFIG_FILE_PATH=test/jobs/dbm-02-drop-defaults.yaml go run ./jobs/db-migration/*.go
CONFIG_FILE_PATH=test/jobs/dbm-03-create-defaults.yaml go run ./jobs/db-migration/*.go
```

Test steps:

- Phase 1 (`dbm-01-before.yaml`): baseline index snapshot exported.
- Manual step: added at least one custom index directly in MongoDB.
- Phase 2 (`dbm-02-drop-defaults.yaml`): dropped default indexes.
- Phase 3 (`dbm-03-create-defaults.yaml`): recreated default indexes.

Result summary:

- Default indexes were removed in phase 2 and recreated in phase 3 as expected.
- Manually added custom index remained untouched by `drop defaults`.

Conclusion:

- The previous v1 behavior is preserved under v2:
  - `drop defaults` removes only default indexes.
  - custom manually added indexes remain untouched.
  - `create defaults` restores the default indexes again.

## Manual Test Protocol (SaveEmailTemplate)

Date: 25.03.2026

Scope:

- Verify unchanged behavior for `SaveEmailTemplate` after switching to `options.FindOneAndReplace().SetUpsert(false).SetReturnDocument(options.After)`.
- Covered cases: create, update existing, and study-template flow.
- Not covered in this run: update with valid but non-existing `id` (requires direct API request outside UI).

Test execution:

1. Created new global email templates in UI and saved them.
2. Edited existing global email templates in UI and saved again.
3. Repeated create and update flow for study-scoped email templates.

Expected and observed results:

1. Create path (`id` empty): new template is inserted and returned with generated `id`.
2. Update path (`id` exists): existing template is replaced/updated and returned with same `id`.
3. Study-template behavior: same create/update behavior as global templates.
4. Unique index behavior: no duplicate email templates with identical `messageType` are created (default unique index is enforced).

Conclusion:

- No behavioral change observed in the covered `SaveEmailTemplate` flows.
- Current implementation remains consistent with previous driver-v1 behavior for UI-accessible paths.
