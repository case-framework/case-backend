package studyutils

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createGlobalKey() string {
	n := 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func testProfileIDtoParticipantMethod(t *testing.T, method string, studySecret string) {

	globalKey := createGlobalKey()

	testProfileID := primitive.NewObjectID().Hex()
	testProfileID2 := primitive.NewObjectID().Hex()

	t.Run("same user same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId != pId2 {
			t.Errorf("pIDs don't match")
		}
	})

	t.Run("different users same keys", func(t *testing.T) {
		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID2, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})

	t.Run("same user different study keys", func(t *testing.T) {

		if method == ID_MAPPING_SAME {
			// In case of mapping same, no transformation is done, this test is not working obv.
			t.Skip()
			return
		}

		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret+"different", method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})

	t.Run("same user different global keys", func(t *testing.T) {

		if method == ID_MAPPING_SAME {
			// In case of mapping same, no transformation is done, this test is not working obv.
			t.Skip()
			return
		}

		pId, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		pId2, err := ProfileIDtoParticipantID(testProfileID, globalKey+"different", studySecret, method)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if pId == pId2 {
			t.Errorf("pIDs shouldn't match")
		}
	})
}

func TestProfileIDtoParticipantIDMethods(t *testing.T) {
	studySecret := "this!study.-a.sd"

	t.Run(ID_MAPPING_SHA224, func(t *testing.T) {
		testProfileIDtoParticipantMethod(t, ID_MAPPING_SHA224, studySecret)
	})

	t.Run(ID_MAPPING_SHA224_B64, func(t *testing.T) {
		testProfileIDtoParticipantMethod(t, ID_MAPPING_SHA224_B64, studySecret)
	})

	t.Run(ID_MAPPING_SHA256, func(t *testing.T) {
		testProfileIDtoParticipantMethod(t, ID_MAPPING_SHA256, studySecret)
	})

	t.Run(ID_MAPPING_SHA256_B64, func(t *testing.T) {
		testProfileIDtoParticipantMethod(t, ID_MAPPING_SHA256_B64, studySecret)
	})

	t.Run(ID_MAPPING_AESCTR, func(t *testing.T) {
		testProfileIDtoParticipantMethod(t, ID_MAPPING_AESCTR, studySecret)
	})
	t.Run(ID_MAPPING_SAME, func(t *testing.T) {
		testProfileIDtoParticipantMethod(t, ID_MAPPING_SAME, studySecret)
	})
}

func benchmarkMappingParticipantID(b *testing.B, method string) {
	studySecret := "this!study.-a.sd"
	globalKey := createGlobalKey()
	for n := 0; n < b.N; n++ {
		testProfileID := primitive.NewObjectID().Hex()
		_, err := ProfileIDtoParticipantID(testProfileID, globalKey, studySecret, method)
		if err != nil {
			b.Errorf("unexpected error: %s", err.Error())
			return
		}
	}
}

func BenchmarkMappingSha224(b *testing.B) { benchmarkMappingParticipantID(b, ID_MAPPING_SHA224) }

func BenchmarkMappingSha224b64(b *testing.B) { benchmarkMappingParticipantID(b, ID_MAPPING_SHA224_B64) }

func BenchmarkMappingSha256(b *testing.B) { benchmarkMappingParticipantID(b, ID_MAPPING_SHA256) }

func BenchmarkMappingSha256b64(b *testing.B) { benchmarkMappingParticipantID(b, ID_MAPPING_SHA256_B64) }

func BenchmarkMappingAesctr(b *testing.B) { benchmarkMappingParticipantID(b, ID_MAPPING_AESCTR) }

func BenchmarkMappingSame(b *testing.B) { benchmarkMappingParticipantID(b, ID_MAPPING_SAME) }
