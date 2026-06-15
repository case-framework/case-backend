# Management API - Logged Events Reference

This document provides a comprehensive list of all logged events in the Management API service. Events are organized by source file and log level to help with monitoring, debugging, and understanding the API's behavior.

The Management API uses Go's standard `log/slog` library for structured logging. Each event includes contextual information such as instance ID, user ID, and relevant resource identifiers.

## Log Levels

- **Info**: Informational messages about normal operations
- **Warn**: Warning messages about potentially problematic situations
- **Error**: Error messages indicating failures that need attention
- **Debug**: Detailed debugging information (usually for development)

---

## handlers.go

### Debug Events

- `Error reading serviceInfos.json` (line 22)
- `Error unmarshalling serviceInfos.json` (line 26)

---

## init.go

### Error Events

- `Error connecting to Global Infos DB` (line 147)
- `Error connecting to Management User DB` (line 123)
- `Error connecting to Messaging DB` (line 129)
- `Error connecting to Participant User DB` (line 141)
- `Error connecting to Study DB` (line 135)
- `Filestore path does not exist` (line 252)
- `Filestore path not set` (line 247)
- `error during initConfig` (line 212)

---

## main.go

### Info Events

- `Starting Management API` (line 58)

### Error Events

- `Error loading TLS config.` (line 69)
- `Exited Management API` (line 62)
- `Exited Management API` (line 81)

---

## management-auth.go

### Info Events

- `extended session` (line 203)
- `getting user permissions` (line 244)
- `got renew token` (line 235)
- `sign in with an existing management user` (line 98)
- `sign up with a new management user` (line 81)

### Warn Events

- `instance not allowed` (line 59)
- `instance not allowed` (line 170)
- `no sessionID` (line 217)
- `no sub` (line 65)
- `user not allowed to get renew token` (line 230)

### Error Events

- `could not create new user` (line 93)
- `could not create session` (line 123)
- `could not create session` (line 181)
- `could not generate token` (line 140)
- `could not generate token` (line 198)
- `could not update existing user` (line 111)
- `error retrieving user app roles` (line 255)
- `error retrieving user permissions` (line 248)
- `failed to bind request` (line 53)
- `failed to bind request` (line 164)

### Debug Events

- `could not get session` (line 225)

---

## messaging-service.go

### Info Events

- `deleting global message template` (line 273)
- `deleting scheduled email` (line 532)
- `deleting study message template` (line 443)
- `getting SMS template` (line 289)
- `getting global message template` (line 248)
- `getting global message templates` (line 202)
- `getting scheduled email` (line 517)
- `getting scheduled emails` (line 458)
- `getting study message template` (line 427)
- `getting study message templates` (line 356)
- `getting study message templates` (line 341)
- `saving SMS template` (line 328)
- `saving global message template` (line 233)
- `saving scheduled email` (line 480)
- `saving study message template` (line 411)

### Error Events

- `error deleting global message template` (line 277)
- `error deleting scheduled email` (line 536)
- `error deleting study message template` (line 447)
- `error getting SMS template` (line 302)
- `error getting global message template` (line 261)
- `error getting global message templates` (line 206)
- `error getting scheduled email` (line 521)
- `error getting scheduled emails` (line 462)
- `error getting studies` (line 374)
- `error getting study message template` (line 431)
- `error getting study message templates` (line 345)
- `error getting study message templates` (line 360)
- `error parsing request body` (line 395)
- `error parsing request body` (line 219)
- `error parsing request body` (line 316)
- `error parsing request body` (line 475)
- `error parsing template` (line 406)
- `error parsing template` (line 323)
- `error parsing template` (line 228)
- `error parsing template` (line 485)
- `error saving SMS template` (line 332)
- `error saving global message template` (line 237)
- `error saving scheduled email` (line 493)
- `error saving scheduled email` (line 498)
- `error saving scheduled email` (line 506)
- `error saving study message template` (line 415)
- `study not found` (line 387)

---

## participant-management.go

### Info Events

- `creating virtual participant` (line 127)
- `getting responses for participant` (line 191)
- `submitting report for participant` (line 250)
- `submitting response for participant` (line 155)
- `updating participant` (line 335)
- `updating report for participant` (line 284)

### Error Events

- `failed to bind request` (line 276)
- `failed to bind request` (line 150)
- `failed to bind request` (line 339)
- `failed to bind request` (line 216)
- `failed to bind request` (line 239)
- `failed to bind request` (line 308)
- `failed to create virtual participant` (line 131)
- `failed to get participant` (line 168)
- `failed to get responses` (line 195)
- `failed to merge participants` (line 321)
- `failed to parse paginated query` (line 184)
- `failed to save report` (line 254)
- `failed to submit event` (line 223)
- `failed to submit response` (line 161)
- `failed to update participant` (line 352)
- `failed to update report` (line 288)
- `participant ID in request does not match participant ID in path` (line 245)
- `participant ID in request does not match participant ID in path` (line 345)
- `target participant ID and with participant ID are the same` (line 314)

---

## study-management.go

### Info Events

- `adding new study code list entries` (line 1905)
- `adding study permission` (line 1721)
- `creating new study` (line 1095)
- `creating study variable` (line 2120)
- `creating survey` (line 1460)
- `deleting study` (line 1409)
- `deleting study code list (full)` (line 1941)
- `deleting study code list entry` (line 1949)
- `deleting study file` (line 4157)
- `deleting study permission` (line 1753)
- `deleting study response` (line 3869)
- `deleting study responses` (line 3845)
- `deleting study rule version` (line 2318)
- `deleting study variable` (line 2211)
- `deleting survey version` (line 1624)
- `downloading daily export file` (line 3635)
- `downloading prepared confidential response export file` (line 3484)
- `exporting study config` (line 1193)
- `fetching available report keys` (line 3960)
- `generating participants export` (line 3038)
- `generating reports export` (line 3223)
- `generating responses export` (line 2823)
- `getting all studies` (line 1061)
- `getting available confidential response exports` (line 3426)
- `getting confidential responses` (line 3383)
- `getting confidential responses for participant` (line 3405)
- `getting current study rules` (line 2228)
- `getting daily exports` (line 3577)
- `getting export task result` (line 3535)
- `getting export task result` (line 2540)
- `getting export task status` (line 3512)
- `getting latest survey` (line 1505)
- `getting notification subscriptions` (line 1783)
- `getting participants count` (line 3007)
- `getting reports count` (line 3194)
- `getting responses count` (line 2794)
- `getting study action task status` (line 2517)
- `getting study code list entries` (line 1864)
- `getting study code list keys` (line 1828)
- `getting study counter values` (line 1966)
- `getting study file` (line 4127)
- `getting study files` (line 4099)
- `getting study participant` (line 3920)
- `getting study participants` (line 3885)
- `getting study props` (line 1144)
- `getting study report` (line 4075)
- `getting study reports` (line 3998)
- `getting study response by ID` (line 3752)
- `getting study responses` (line 3677)
- `getting study rule version` (line 2299)
- `getting study rule versions` (line 2280)
- `getting study variable` (line 2085)
- `getting study variables` (line 2062)
- `getting survey info` (line 2740)
- `getting survey info list` (line 1430)
- `getting survey version` (line 1604)
- `getting survey versions` (line 1585)
- `incrementing study counter` (line 2022)
- `publishing new study rules version` (line 2263)
- `removing study counter` (line 2046)
- `running study action on participant` (line 2344)
- `running study action on participants` (line 2646)
- `running study action on participants` (line 2445)
- `running study action on previous responses for participant` (line 2607)
- `saving study counter value` (line 1998)
- `unpublishing survey` (line 1567)
- `updating notification subscriptions` (line 1811)
- `updating study display props` (line 1344)
- `updating study file upload rule` (line 1393)
- `updating study is default` (line 1287)
- `updating study status` (line 1315)
- `updating study variable definition` (line 2154)
- `updating study variable value` (line 2188)
- `updating survey` (line 1549)

### Warn Events

- `permission does not belong to the study` (line 1763)
- `user is not allowed to get task result` (line 2550)
- `user is not allowed to get task result` (line 3545)
- `user is not allowed to get task status` (line 2527)
- `user is not allowed to get task status` (line 3522)

### Error Events

- `Error deleting study code list` (line 1944)
- `Error deleting study code list entry` (line 1953)
- `Missing required parameters` (line 1934)
- `controlField does not match studyKey` (line 3830)
- `could not write key into config export` (line 1176)
- `could not write simple string into config export` (line 1163)
- `could not write value into config export` (line 1180)
- `error decoding exportID` (line 3479)
- `error decoding exportID` (line 3630)
- `error parsing fromTS` (line 3943)
- `error parsing fromTS` (line 4021)
- `error parsing toTS` (line 3954)
- `error parsing toTS` (line 4032)
- `error retrieving unique report keys` (line 3984)
- `failed to add study code list entry` (line 1917)
- `failed to add study permission` (line 1738)
- `failed to bind request` (line 3372)
- `failed to bind request` (line 2183)
- `failed to bind request` (line 1716)
- `failed to bind request` (line 1367)
- `failed to bind request` (line 2641)
- `failed to bind request` (line 2602)
- `failed to bind request` (line 2440)
- `failed to bind request` (line 1454)
- `failed to bind request` (line 1090)
- `failed to bind request` (line 2339)
- `failed to bind request` (line 2247)
- `failed to bind request` (line 1339)
- `failed to bind request` (line 2149)
- `failed to bind request` (line 1525)
- `failed to bind request` (line 2113)
- `failed to bind request` (line 1993)
- `failed to bind request` (line 1310)
- `failed to bind request` (line 1282)
- `failed to bind request` (line 1806)
- `failed to bind request` (line 1893)
- `failed to convert response to flat object` (line 3808)
- `failed to convert response to flat object` (line 3734)
- `failed to create action run results file` (line 2404)
- `failed to create actionRuns folder` (line 2651)
- `failed to create actionRuns folder` (line 2450)
- `failed to create export file` (line 3266)
- `failed to create export file` (line 3082)
- `failed to create export file` (line 2907)
- `failed to create export folder` (line 3255)
- `failed to create export folder` (line 3070)
- `failed to create export folder` (line 2891)
- `failed to create export task` (line 3062)
- `failed to create export task` (line 3247)
- `failed to create export task` (line 2883)
- `failed to create response exporter` (line 2921)
- `failed to create response parser` (line 2865)
- `failed to create response parser` (line 3719)
- `failed to create response parser` (line 3794)
- `failed to create study` (line 1131)
- `failed to create study variable` (line 2124)
- `failed to create survey` (line 1491)
- `failed to create task` (line 2462)
- `failed to create task` (line 2663)
- `failed to decode JSON file` (line 2582)
- `failed to delete study` (line 1413)
- `failed to delete study file` (line 4169)
- `failed to delete study file` (line 4181)
- `failed to delete study file preview` (line 4174)
- `failed to delete study permission` (line 1770)
- `failed to delete study response` (line 3873)
- `failed to delete study responses` (line 3849)
- `failed to delete study rule version` (line 2322)
- `failed to delete study variable` (line 2215)
- `failed to delete survey version` (line 1628)
- `failed to export participants` (line 3148)
- `failed to export reports` (line 3331)
- `failed to export responses` (line 2965)
- `failed to finish export` (line 2972)
- `failed to get all studies` (line 1065)
- `failed to get confidential participantID` (line 3402)
- `failed to get confidential responses` (line 3409)
- `failed to get current study rules` (line 2232)
- `failed to get export task result` (line 2544)
- `failed to get export task result` (line 3539)
- `failed to get export task status` (line 2521)
- `failed to get export task status` (line 3516)
- `failed to get latest survey` (line 1253)
- `failed to get latest survey` (line 1509)
- `failed to get notification subscriptions` (line 1787)
- `failed to get participants count` (line 3042)
- `failed to get participants count` (line 3011)
- `failed to get reports count` (line 3227)
- `failed to get reports count` (line 3198)
- `failed to get responses count` (line 2827)
- `failed to get responses count` (line 2798)
- `failed to get rules for study` (line 1242)
- `failed to get study` (line 1204)
- `failed to get study` (line 1148)
- `failed to get study` (line 3387)
- `failed to get study code list entries` (line 1868)
- `failed to get study code list keys` (line 1832)
- `failed to get study counter values` (line 1970)
- `failed to get study file info` (line 4131)
- `failed to get study file info` (line 4161)
- `failed to get study files` (line 4109)
- `failed to get study participant` (line 3924)
- `failed to get study participants` (line 3903)
- `failed to get study permission` (line 1757)
- `failed to get study permissions` (line 1648)
- `failed to get study report` (line 4079)
- `failed to get study reports` (line 4058)
- `failed to get study response by ID` (line 3763)
- `failed to get study responses` (line 3688)
- `failed to get study rule version` (line 2303)
- `failed to get study rule versions` (line 2284)
- `failed to get study variable` (line 2089)
- `failed to get study variables` (line 2066)
- `failed to get survey info` (line 2752)
- `failed to get survey info csv` (line 2774)
- `failed to get survey info list` (line 1434)
- `failed to get survey info list` (line 1464)
- `failed to get survey infos for study` (line 1263)
- `failed to get survey version` (line 1609)
- `failed to get survey versions` (line 2851)
- `failed to get survey versions` (line 1480)
- `failed to get survey versions` (line 1540)
- `failed to get survey versions` (line 1589)
- `failed to get survey versions` (line 3705)
- `failed to get survey versions` (line 3780)
- `failed to get user info` (line 1687)
- `failed to increment study counter` (line 2026)
- `failed to marshal participant` (line 3121)
- `failed to marshal report` (line 3304)
- `failed to marshal study rules` (line 2258)
- `failed to open file` (line 2573)
- `failed to parse filter` (line 3002)
- `failed to parse filter` (line 2787)
- `failed to parse filter` (line 3026)
- `failed to parse filter` (line 3213)
- `failed to parse filter` (line 3184)
- `failed to parse paginated query` (line 3889)
- `failed to parse paginated query` (line 4002)
- `failed to parse query` (line 4094)
- `failed to parse query` (line 2812)
- `failed to parse query` (line 3665)
- `failed to parse query` (line 1848)
- `failed to parse query` (line 1854)
- `failed to parse response` (line 3729)
- `failed to parse response` (line 3801)
- `failed to parse response export query` (line 3823)
- `failed to parse response export query` (line 3756)
- `failed to parse sort` (line 3033)
- `failed to publish new study rules version` (line 2267)
- `failed to remove study counter` (line 2050)
- `failed to run study action` (line 2355)
- `failed to run study action` (line 2618)
- `failed to run study actions` (line 2394)
- `failed to save study counter value` (line 2002)
- `failed to unpublish survey` (line 1571)
- `failed to update notification subscriptions` (line 1815)
- `failed to update study display props` (line 1348)
- `failed to update study file upload rule` (line 1397)
- `failed to update study is default` (line 1291)
- `failed to update study status` (line 1319)
- `failed to update study variable definition` (line 2158)
- `failed to update study variable value` (line 2192)
- `failed to update survey` (line 1553)
- `failed to update task progress` (line 3321)
- `failed to update task progress` (line 2494)
- `failed to update task progress` (line 2953)
- `failed to update task progress` (line 3138)
- `failed to update task progress` (line 2696)
- `failed to update task status` (line 2427)
- `failed to update task status` (line 3352)
- `failed to update task status` (line 2986)
- `failed to update task status` (line 3169)
- `failed to update task status on faied task` (line 2381)
- `failed to update task total count` (line 2684)
- `failed to update task total count` (line 2482)
- `failed to write footer` (line 3155)
- `failed to write footer` (line 3338)
- `failed to write header` (line 3092)
- `failed to write header` (line 3276)
- `failed to write to action run results file` (line 2413)
- `failed to write to export file` (line 3309)
- `failed to write to export file` (line 3296)
- `failed to write to export file` (line 3113)
- `failed to write to export file` (line 3126)
- `file does not exist` (line 2565)
- `file does not exist` (line 4140)
- `file does not exist` (line 3560)
- `file does not exist` (line 3641)
- `file does not exist` (line 3490)
- `invalid format` (line 2732)
- `list key is empty` (line 1900)
- `participantIDs is required` (line 3378)
- `responseID is required` (line 3864)
- `running study actions resulted in error` (line 2707)
- `running study actions resulted in error` (line 2501)
- `scope is required` (line 2017)
- `scope is required` (line 2041)
- `scope is required` (line 1984)
- `secret key is too short` (line 1105)
- `study key is not URL safe` (line 1099)
- `studyKey and variableKey are required` (line 2142)
- `studyKey and variableKey are required` (line 2176)
- `studyKey and variableKey are required` (line 2080)
- `studyKey and variableKey are required` (line 2206)
- `studyKey is required` (line 2106)
- `survey key already exists` (line 1471)
- `survey key in request does not match` (line 1532)
- `surveyKey is required` (line 3837)
- `surveyKey is required` (line 2818)
- `surveyKey is required` (line 3672)
- `surveyKey is required` (line 2725)
- `task is not completed` (line 3551)
- `task is not completed` (line 2556)
- `unexpected error when reading confidential file exports` (line 3460)
- `unexpected error when reading daily file exports` (line 3611)

---

## user-management.go

### Info Events

- `adding user app role` (line 290)
- `creating a new service account` (line 617)
- `creating app role template` (line 375)
- `creating service account API key` (line 698)
- `creating service account permission` (line 785)
- `creating user permission` (line 187)
- `deleting all app role templates (and its instance roles) for app` (line 485)
- `deleting all app roles for app` (line 527)
- `deleting app role template` (line 463)
- `deleting service account` (line 748)
- `deleting service account API key` (line 733)
- `deleting service account permission` (line 821)
- `deleting user` (line 135)
- `deleting user app role` (line 339)
- `deleting user permission` (line 222)
- `getting API keys for service account` (line 669)
- `getting all app role templates` (line 354)
- `getting all app roles` (line 505)
- `getting all service accounts` (line 590)
- `getting all users` (line 97)
- `getting app role template by ID` (line 410)
- `getting service account` (line 633)
- `getting user` (line 113)
- `getting user app roles` (line 262)
- `getting user permissions` (line 762)
- `getting user permissions` (line 164)
- `requesting participant user deletion` (line 557)
- `updating app role template` (line 440)
- `updating service account` (line 655)
- `updating service account permission limiter` (line 846)
- `updating user permission limiter` (line 246)

### Error Events

- `app name is required` (line 480)
- `app name is required` (line 522)
- `app role template ID is required` (line 458)
- `app role template ID is required` (line 405)
- `app role template ID is required` (line 428)
- `cannot delete user` (line 578)
- `error adding user app role` (line 309)
- `error binding permission` (line 182)
- `error binding permission` (line 780)
- `error binding permission` (line 841)
- `error binding permission` (line 241)
- `error creating app role template` (line 391)
- `error creating service account` (line 621)
- `error creating service account permission` (line 807)
- `error creating user permission` (line 209)
- `error creating user permission` (line 325)
- `error deleting app role template` (line 467)
- `error deleting app role template for app` (line 494)
- `error deleting app roles for app` (line 531)
- `error deleting app roles for app` (line 488)
- `error deleting permissions` (line 146)
- `error deleting service account permission` (line 825)
- `error deleting sessions` (line 140)
- `error deleting user` (line 152)
- `error deleting user app role` (line 343)
- `error deleting user permission` (line 226)
- `error retrieving app role template` (line 302)
- `error retrieving app role template` (line 414)
- `error retrieving app role templates` (line 358)
- `error retrieving app roles` (line 509)
- `error retrieving sercice account permissions` (line 766)
- `error retrieving service account` (line 637)
- `error retrieving service accounts` (line 594)
- `error retrieving user` (line 117)
- `error retrieving user app roles` (line 266)
- `error retrieving user permissions` (line 316)
- `error retrieving user permissions` (line 168)
- `error retrieving users` (line 101)
- `error updating app role template` (line 444)
- `error updating service account permission limiter` (line 850)
- `error updating user permission limiter` (line 250)
- `failed to bind request` (line 546)
- `failed to bind request` (line 693)
- `failed to bind request` (line 379)
- `failed to bind request` (line 612)
- `failed to bind request` (line 435)
- `failed to bind request` (line 650)
- `failed to create service account API key` (line 721)
- `failed to delete service account` (line 751)
- `failed to delete service account API key` (line 736)
- `failed to delete temp tokens` (line 573)
- `failed to generate unique token string` (line 715)
- `failed to get api keys for service account` (line 673)
- `failed to update service account` (line 658)
- `invalid email format` (line 552)
- `service account not found` (line 702)
- `service account not found` (line 789)
- `user cannot delete itself` (line 130)
- `user not found` (line 191)
- `user not found` (line 561)
- `user not found` (line 294)

---

## utils.go

### Warn Events

- `unauthorised access attempted` (line 66)

### Error Events

- `failed to update task status` (line 95)

---

## Summary

This document lists **350 unique log events** across **9 source files** in the Management API service. The events are categorized into:

- **Info Events**: Normal operational events (primary focus for monitoring workflow)
- **Warn Events**: Warnings about potentially problematic situations (authentication failures, validation errors)
- **Error Events**: Errors requiring attention (database failures, parsing errors)
- **Debug Events**: Detailed debugging information (typically for development)

### Usage

When monitoring or debugging the Management API:

1. **Info events** help track normal operations and user activities
2. **Warn events** indicate potential security issues or invalid requests
3. **Error events** require investigation and potential fixes
4. **Debug events** provide additional context during troubleshooting

### Contextual Information

Most log events include contextual attributes such as:
- `instanceID`: The instance being accessed
- `userID` or `Subject`: The user performing the action
- `studyKey`: The study being operated on
- `error`: Error details when operations fail
- Resource-specific identifiers (participantID, surveyKey, etc.)

This structured logging enables effective filtering and analysis in log aggregation systems.
