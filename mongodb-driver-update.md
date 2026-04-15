# Update mongodb driver

- The bson/primitive package has been merged into the bson package —> changed any instance of primitive.ObjectID to bson.ObjectId.

## All DB packages

- context.Context parameter has been removed from mongo.Connect() because the deployment connector doesn’t accept a context, meaning that the context passed to mongo.Connect() in previous versions didn't serve a purpose.
- Simplfied DropOne and DropAll methods by removing the server response
- removed unused return value for context.WithTimeout()
- Index Model: The old `IndexOptionsBuilder` type was removed and `IndexModel.Options.Name` is no longer accessible as a field. Required steps:
  - define static default index name constants/lists per collection
  - reuse those names for both index creation and `DropOne` in `drop defaults`
  - TESTING REQUIRED: expired Indexes and unique indexes

### Messaging

#### email-templates

- save email template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (manually tested, see SaveEmailTemplate protocol)

#### scheduled-emails

- save scheduled-emails: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (manually tested)

#### sms-templates

- save sms template: use the options builder, and pass a pointer returned by options.FindOneAndReplace() instead of constructing FindOneAndReplaceOptions as a struct literal (manually tested, see SaveScheduledEmail protocol)

### participant user

#### user-attributes

- The UpdateOptions has been changed to UpdateOneOptions to configure UpdateOne operation. (manually tested, see `SetUserAttribute` protocol)

#### users

- add user: MongoDB Go Driver v2 no longer allows constructing or modifying option structs directly, so update options must now be created through the new builder API (options.UpdateOne().SetUpsert(true)) instead of setting fields on UpdateOptions manually.
- update user in db: FindOneAndReplaceOptions can no longer be created or populated as a struct, so the v2 driver requires using the builder pattern (options.FindOneAndReplace().SetReturnDocument(options.After)) instead of setting option fields directly. (manually tested, see `_updateUserInDB` protocol)

#### otps

- update the callback for mongo.WithSession to use a context.Context implementation, rather than the custom mongo.SessionContext TESTING REQUIRED

TEST: If you want to be extra safe, you can:
Deploy this change to a non‑production environment first.
Run a few test flows that:
exceed the maxOTPCount to ensure the “too many OTP requests” path still works,
run concurrent OTP creations to confirm only the allowed number of documents is written.

### study

#### participants

- saveParticipantSate: configure FindOneAndReplaceOptions through options.FindOneAndReplace().Set... instead of filling the struct fields directly.

#### confidential responses

- configure replace options via the options.Replace() builder (for example, options.Replace().SetUpsert(true)) instead of instantiating a ReplaceOptions struct literal and passing its address.

#### study-rules

- &options.FindOneOptions{ Sort: sortByPublished } becomes options.FindOne().SetSort(sortByPublished).

#### reports

- GetUniqueReportKeysForStudy:`Distinct()` no longer returns `([]interface{}, error)`; it returns a single result type on which you call `.Decode(&target)` directly into a `[]string`, eliminating the manual type-assertion loop. (manually tested, see `GetUniqueReportKeysForStudy` protocol)

#### surveys

- GetSurveyKeysForStudy: `Distinct()` no longer returns `([]interface{}, error)`; it returns a single result type on which you call `.Decode(&target)` directly into a `[]string`, eliminating the manual type-assertion loop. (manually tested, see `GetSurveyKeysForStudy` protocol)
- GetCurrentSurveyVersion: create FindOneOptions using the options.FindOne() builder and setters (for example, options.FindOne().SetSort(sortByPublishedDesc)) instead of instantiating &options.FindOneOptions{} and mutating its fields.
- GetSurveyVersions: create FindOptions using the options.Find() builder and its setters (for example, options.Find().SetProjection(...).SetSort(...)) instead of instantiating &options.FindOptions{} and mutating its fields.

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

## Manual Test Protocol (SaveScheduledEmail)

Date: 07.04.2026

Scope:

- Verify unchanged behavior for scheduled-email and sms-template save flows after switching to options-builder usage for FindOneAndReplace options.

Test execution:

1. Created new scheduled emails in UI and saved them.
2. Updated existing scheduled emails in UI and saved the changes.
3. Created new SMS templates in UI and saved them.
4. Updated existing SMS templates in UI and saved the changes.

Expected and observed results:

1. Create path: new schedules are created successfully.
2. Update path: existing schedules are updated successfully.
3. Create/update path for SMS templates: works successfully as before.

Conclusion:

- No behavioral change observed in the tested `SaveScheduledEmail` and `SaveSMSTemplate` create/update flows.

## Manual Test Protocol (GetSurveyVersions)

Date: 09.04.2026

Scope:

- Verify unchanged behavior for `GetSurveyVersions` after switching to `options.Find()` builder with setters.

Test execution:

1. Opened survey version list in UI for a study/survey.
2. Verified API request used endpoint `GET /v1/studies/:studyKey/surveys/:surveyKey/versions`.

Expected and observed results:

1. Survey versions are returned and displayed as expected.
2. No behavioral change observed compared to previous behavior.

Conclusion:

- `GetSurveyVersions` behavior is unchanged for the tested UI/API flow.

## Manual Test Protocol (GetCurrentSurveyVersions)

Date: 10.04.2026

Scope:

- Verify unchanged behavior for `GetCurrentSurveyVersions` after switching to `options.FindOne()` builder with setters.

Test execution:

1. Exported study configuration as JSON via the UI (study configuration export).
2. `GetCurrentSurveyVersions` is called internally during this export to retrieve the current version of all surveys belonging to the study.

Expected and observed results:

1. The exported JSON file is identical to the file exported with the previous MongoDB driver version.
2. No behavioral change observed.

Conclusion:

- `GetCurrentSurveyVersions` behavior is unchanged after the driver migration.

## Manual Test Protocol (GetCurrentStudyRules)

Date: 10.04.2026

Scope:

- Verify unchanged behavior for `GetCurrentStudyRules` after switching to `options.FindOne()` builder with setters.

Test execution:

1. Exported study configuration as JSON via the UI (study configuration export).
2. `GetCurrentStudyRules` is called internally during this export to retrieve the current study rules belonging to the study.

Expected and observed results:

1. The exported JSON file is identical to the file exported with the previous MongoDB driver version.
2. No behavioral change observed.

Conclusion:

- `GetCurrentStudyRules` behavior is unchanged after the driver migration.

## Manual Test Protocol (ReplaceConfidentialResponse)

Date: 13.04.2026

Scope:

- Verify unchanged behavior for `ReplaceConfidentialResponse` after switching to `options.Replace().SetUpsert(true)` builder usage.

Test execution:

1. Submitted a survey with confidential responses for a participant for the first time (insert path).
2. Submitted the same survey again for the same participant (replace path), so the existing confidential response was replaced by the new one.

Expected and observed results:

1. First submission: confidential response was inserted as a new document in the database.
2. Subsequent submission: existing confidential response was replaced by the new response, no duplicate was created.

Conclusion:

- No behavioral change observed for the insert and replace paths of `ReplaceConfidentialResponse`.

## Manual Test Protocol (SaveParticipantState)

Date: 13.04.2026

Scope:

- Verify unchanged behavior for `SaveParticipantState` after switching to `options.FindOneAndReplace().SetUpsert(true).SetReturnDocument(options.After)` builder usage.

Test execution:

1. Created a new participant by enrolling in a study (insert path).
2. Changed the state of an existing participant (replace path), e.g. by submitting a survey that modifies participant flags or survey assignments.

Expected and observed results:

1. Insert path: new participant document was created successfully in the database.
2. Replace path: existing participant state was updated correctly, no duplicate was created.

Conclusion:

- No behavioral change observed for the insert and replace paths of `SaveParticipantState`.

## Manual Test Protocol (AddUser)

Date: 14.04.2026

Scope:

- Verify unchanged behavior for `AddUser` after switching to `options.UpdateOne().SetUpsert(true)` builder usage.

Test execution:

1. Created a new user with a fresh email address (insert path).
2. Attempted to create another new user with the same email address as an already existing user (duplicate path).

Expected and observed results:

1. Insert path: new user was created successfully in the database.
2. Duplicate path: error `"user already exists"` was returned correctly, no duplicate user was created in the database.

Conclusion:

- No behavioral change observed for the insert and duplicate paths of `AddUser`.

## Manual Test Protocol (_updateUserInDB)

Date: 14.04.2026

Scope:

- Verify unchanged behavior for `_updateUserInDB` (called via `ReplaceUser`) after switching to `options.FindOneAndReplace().SetReturnDocument(options.After)` builder usage.

Test execution:

1. Logged in as a participant user (triggers `ReplaceUser` → `_updateUserInDB` to update login timestamps).
2. Created a new profile for the logged-in user (triggers another `ReplaceUser` → `_updateUserInDB` to persist the updated user document).

Expected and observed results:

1. Login: user document was updated correctly in the database (e.g. `lastLogin` timestamp).
2. New profile: profile was saved correctly and the updated user document was returned, no duplicate was created.

Conclusion:

- No behavioral change observed for the replace path of `_updateUserInDB`.

## Manual Test Protocol (SetUserAttribute)

Date: 15.04.2026

Scope:

- Verify unchanged behavior for `SetUserAttribute` after switching to `options.UpdateOne().SetUpsert(true)` builder usage.

Test execution:

1. Registered a new user at flusurvey (insert path: user attribute document created for the first time).

Expected and observed results:

1. Insert path: user attribute was created successfully as a new document in the database.

Conclusion:

- No behavioral change observed for the insert path of `SetUserAttribute`.

## Manual Test Protocol (GetUniqueReportKeysForStudy)

Date: 15.04.2026

Scope:

- Verify unchanged behavior for `GetUniqueReportKeysForStudy` after switching to the new `Distinct()` API that returns a result type decoded directly into `[]string`.

Test execution:

1. Opened the reports page in the CASE management UI for a study with existing reports.
2. Checked the dropdown that lists all available report keys.
3. Applied a `participantID` filter and verified the dropdown updated accordingly.
4. Applied `from` and `until` date filters and verified the dropdown updated accordingly.

Expected and observed results:

1. All available and filtered report keys were listed correctly in the dropdown.
2. No behavioral change observed compared to previous behavior.

Conclusion:

- No behavioral change observed for `GetUniqueReportKeysForStudy` after the driver migration.

## Manual Test Protocol (GetSurveyKeysForStudy)

Date: 15.04.2026

Scope:

- Verify unchanged behavior for `GetSurveyKeysForStudy` after switching to the new `Distinct()` API that returns a result type decoded directly into `[]string`.

Test execution:

1. Opened the CASE management UI and checked the survey dropdown used to manually assign a survey to a participant – verified that all available survey keys were listed correctly.
2. Exported study configuration as JSON via the UI and verified that the survey keys are correctly included in the export.

Expected and observed results:

1. All available survey keys were listed correctly in the dropdown.
2. The exported study configuration JSON contained the correct survey keys, identical to the export with the previous MongoDB driver version.

Conclusion:

- No behavioral change observed for `GetSurveyKeysForStudy` after the driver migration.
