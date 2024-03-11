package surveyresponses

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	sd "github.com/case-framework/case-backend/pkg/exporter/survey-definition"
)

func findSurveyVersion(
	versionID string,
	submittedAt int64,
	versions []sd.SurveyVersionPreview,
) (sd.SurveyVersionPreview, error) {
	if versionID == "" {
		return findVersionBasedOnTimestamp(submittedAt, versions)
	} else {
		sv, err := findVersionBasedOnVersionID(versionID, versions)
		if err != nil {
			return findVersionBasedOnTimestamp(submittedAt, versions)
		}
		return sv, nil
	}
}

func findVersionBasedOnTimestamp(submittedAt int64, versions []sd.SurveyVersionPreview) (sv sd.SurveyVersionPreview, err error) {
	nearestTime := time.Now().Unix()
	var preVersion sd.SurveyVersionPreview
	//search version with nearest published time < submittedAt
	for _, v := range versions {
		if v.Published <= submittedAt && submittedAt-v.Published <= nearestTime {
			nearestTime = submittedAt - v.Published
			preVersion = v
		}
	}
	if preVersion.Published != 0 {
		if preVersion.Unpublished == 0 || preVersion.Unpublished >= submittedAt {
			return preVersion, nil
		}
	}
	//search version with nearest published time > submittedAt
	nearestTime = time.Now().Unix()
	var postVersion sd.SurveyVersionPreview
	for _, v := range versions {
		if v.Published >= submittedAt && v.Published-submittedAt <= nearestTime {
			nearestTime = v.Published - submittedAt
			postVersion = v
		}
	}
	if postVersion.Published != 0 {
		slog.Debug("Version not found, taking more recent version.")
		return postVersion, nil
	}
	if preVersion.Published != 0 {
		slog.Debug("Version not found, no recent version found, taking older version")
		return preVersion, nil
	}
	return sv, fmt.Errorf("no survey version found: %d", submittedAt)
}

func findVersionBasedOnVersionID(versionID string, versions []sd.SurveyVersionPreview) (sv sd.SurveyVersionPreview, err error) {
	for _, v := range versions {
		if v.VersionID == versionID {
			return v, nil
		}
	}
	return sv, errors.New("no survey version found")
}
