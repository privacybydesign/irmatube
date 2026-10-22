package main

import "strconv"

const (
	typeAttribute = "type"
	idAttribute   = "id"

	memberTypeRegular = "regular"
	memberTypePremium = "premium"
)

// ageAttributePrefixes are the issuers of "over N" attributes that
// YiviTube accepts as proof of age.
var ageAttributePrefixes = []string{
	"pbdf.pbdf.ageLimits.over",
	"pbdf.nijmegen.ageLimits.over",
	"pbdf.gemeente.personalData.over",
	"pbdf.pilot-amsterdam.idcard.over",
	"pbdf.pilot-amsterdam.passport.over",
}

func membershipAttributeID(credentialID string) string {
	return credentialID + "." + typeAttribute
}

func ageAttributeIDs(age int) []string {
	ids := make([]string, 0, len(ageAttributePrefixes))
	for _, prefix := range ageAttributePrefixes {
		ids = append(ids, prefix+strconv.Itoa(age))
	}

	return ids
}
