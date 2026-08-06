package amlcommon

type InquiryResponse struct {
	Code            string              `json:"code"`
	Message         string              `json:"message"`
	Data            InquiryResponseData `json:"data"`
	TransactionID   string              `json:"transactionId"`
	PricingStrategy string              `json:"pricingStrategy"`
	Datetime        float64             `json:"datetime"`
	Timestamp       int64               `json:"timestamp"`
}

type InquiryResponseData struct {
	ID              string          `json:"id"`
	Journey         Journey         `json:"journey"`
	CustomerProfile CustomerProfile `json:"customerProfile"`
	Nodes           []Node          `json:"nodes"`
}

type Journey struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CustomerProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Node struct {
	Type               string      `json:"type"`
	Name               string      `json:"name"`
	ID                 int         `json:"id"`
	Code               *string     `json:"code"`
	Message            *string     `json:"message"`
	StartedAt          string      `json:"startedAt"`
	CompletedAt        string      `json:"completedAt"`
	Attributes         any         `json:"attributes"`
	Result             *NodeResult `json:"result"`
	VerificationResult string      `json:"verificationResult"`
}

type NodeResult struct {
	Detail       []NodeDetail `json:"detail"`
	MatchedCount int          `json:"matchedCount"`
	Summary      NodeSummary  `json:"summary"`
}

type NodeDetail struct {
	AliasName        []string `json:"aliasName"`
	CreateTime       float64  `json:"createTime"`
	CreationDate     string   `json:"creationDate"`
	DateOfBirth      string   `json:"dateOfBirth"`
	HitCategory      []string `json:"hitCategory"`
	LastAlertDate    string   `json:"lastAlertDate"`
	MatchScore       int      `json:"matchScore"`
	MatchStatus      int      `json:"matchStatus"`
	ModificationDate string   `json:"modificationDate"`
	Name             string   `json:"name"`
	ProfileID        string   `json:"profileId"`
	Region           []string `json:"region"`
	ResultID         string   `json:"resultId"`
	Type             string   `json:"type"`
}

type NodeSummary struct {
	FullMarksNum int      `json:"fullMarksNum"`
	IconHitSet   []string `json:"iconHitSet"`
	Total        int      `json:"total"`
}

type NodeAttributes struct {
	DOB               string   `json:"dob"`
	Name              string   `json:"name"`
	Score             int      `json:"score"`
	Gender            string   `json:"gender"`
	EntityType        string   `json:"entityType"`
	HitCategory       []string `json:"hitCategory"`
	ReferenceID       string   `json:"referenceId"`
	PlaceOfBirth      string   `json:"placeOfBirth"`
	CountryLocation   string   `json:"countryLocation"`
	RegisteredCountry string   `json:"registeredCountry"`
}

// cant create a struct for this, since attributes may differ
// depending on the journey
func ExtractNodeAttributes(attributes any) NodeAttributes {
	var result NodeAttributes

	if nodeAttrs, ok := attributes.(map[string]any); ok {
		if dob, ok := nodeAttrs["dob"].(string); ok {
			result.DOB = dob
		}
		if name, ok := nodeAttrs["name"].(string); ok {
			result.Name = name
		}
		if score, ok := nodeAttrs["score"].(float64); ok {
			result.Score = int(score)
		}
		if gender, ok := nodeAttrs["gender"].(string); ok {
			result.Gender = gender
		}
		if entityType, ok := nodeAttrs["entityType"].(string); ok {
			result.EntityType = entityType
		}
		if refID, ok := nodeAttrs["referenceId"].(string); ok {
			result.ReferenceID = refID
		}
		if pob, ok := nodeAttrs["placeOfBirth"].(string); ok {
			result.PlaceOfBirth = pob
		}
		if country, ok := nodeAttrs["countryLocation"].(string); ok {
			result.CountryLocation = country
		}
		if regCountry, ok := nodeAttrs["registeredCountry"].(string); ok {
			result.RegisteredCountry = regCountry
		}
		if hitCat, ok := nodeAttrs["hitCategory"].([]any); ok {
			result.HitCategory = make([]string, len(hitCat))
			for i, cat := range hitCat {
				if catStr, ok := cat.(string); ok {
					result.HitCategory[i] = catStr
				}
			}
		}
	}

	return result
}
