package amlcommon

type ProfileDetailResponse struct {
	Code            string            `json:"code"`
	Message         string            `json:"message"`
	Data            ProfileDetailData `json:"data"`
	ReferenceID     string            `json:"referenceId"`
	TransactionID   string            `json:"transactionId"`
	Datetime        float64           `json:"datetime"`
	Extra           any               `json:"extra"`
	PricingStrategy string            `json:"pricingStrategy"`
	Timestamp       int64             `json:"timestamp"`
}

type ProfileDetailData struct {
	ProfileID         string            `json:"profileId"`
	ResultID          string            `json:"resultId"`
	Name              string            `json:"name"`
	AliasName         []string          `json:"aliasName"`
	HitCategory       []string          `json:"hitCategory"`
	DateOfBirth       string            `json:"dateOfBirth"`
	CreationDate      string            `json:"creationDate"`
	ModificationDate  string            `json:"modificationDate"`
	LastAlertDate     string            `json:"lastAlertDate"`
	Score             int               `json:"score"`
	ProfileRecordInfo ProfileRecordInfo `json:"profileRecordInfo"`
}

type ProfileRecordInfo struct {
	Actions               []any         `json:"actions"`
	Active                bool          `json:"active"`
	Addresses             []Address     `json:"addresses"`
	Associates            []Associate   `json:"associates"`
	Category              string        `json:"category"`
	Comments              any           `json:"comments"`
	Contacts              []any         `json:"contacts"`
	CountryLinks          []CountryLink `json:"countryLinks"`
	CreationDate          string        `json:"creationDate"`
	DeletionDate          any           `json:"deletionDate"`
	Description           any           `json:"description"`
	Details               []Detail      `json:"details"`
	EntityID              string        `json:"entityId"`
	ExternalImportID      string        `json:"externalImportId"`
	Files                 []any         `json:"files"`
	IdentityDocuments     []any         `json:"identityDocuments"`
	Images                []any         `json:"images"`
	LastAdjunctChangeDate any           `json:"lastAdjunctChangeDate"`
	ModificationDate      string        `json:"modificationDate"`
	Names                 []Name        `json:"names"`
	PreviousCountryLinks  []any         `json:"previousCountryLinks"`
	Provider              Provider      `json:"provider"`
	SourceDescription     any           `json:"sourceDescription"`
	SourceUris            []any         `json:"sourceUris"`
	Sources               []Source      `json:"sources"`
	SubCategory           string        `json:"subCategory"`
	UpdateCategory        string        `json:"updateCategory"`
	UpdatedDates          UpdatedDates  `json:"updatedDates"`
	Weblinks              []Weblink     `json:"weblinks"`
	TranslatedEntity      any           `json:"translatedEntity"`
	Age                   any           `json:"age"`
	AgeAsOfDate           any           `json:"ageAsOfDate"`
	Events                []Event       `json:"events"`
	Gender                string        `json:"gender"`
	IsDeceased            bool          `json:"isDeceased"`
	PreviousRoles         []any         `json:"previousRoles"`
	Roles                 []Role        `json:"roles"`
	EntityType            string        `json:"entityType"`
}

type Address struct {
	City     string  `json:"city"`
	Country  Country `json:"country"`
	PostCode any     `json:"postCode"`
	Region   string  `json:"region"`
	Street   any     `json:"street"`
}

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Associate struct {
	AssociateEntityType    string   `json:"associateEntityType"`
	EntityType             string   `json:"entityType"`
	Reversed               bool     `json:"reversed"`
	TargetEntityID         string   `json:"targetEntityId"`
	TargetExternalImportID string   `json:"targetExternalImportId"`
	Type                   string   `json:"type"`
	TargetPrimaryName      string   `json:"targetPrimaryName"`
	CreationDate           string   `json:"creationDate"`
	ModificationDate       string   `json:"modificationDate"`
	UpdateCategory         string   `json:"updateCategory"`
	Category               string   `json:"category"`
	TargetCategories       []string `json:"targetCategories"`
}

type CountryLink struct {
	Country     Country `json:"country"`
	CountryText string  `json:"countryText"`
	Type        string  `json:"type"`
}

type Detail struct {
	DetailType string `json:"detailType"`
	Text       string `json:"text"`
	Title      string `json:"title"`
}

type Name struct {
	FullName       string `json:"fullName"`
	GivenName      any    `json:"givenName"`
	LanguageCode   any    `json:"languageCode"`
	LastName       any    `json:"lastName"`
	OriginalScript string `json:"originalScript"`
	Prefix         any    `json:"prefix"`
	Suffix         any    `json:"suffix"`
	Type           string `json:"type"`
	Primary        any    `json:"primary"`
}

type Provider struct {
	Code       string `json:"code"`
	Identifier string `json:"identifier"`
	Master     bool   `json:"master"`
	Name       string `json:"name"`
}

type Source struct {
	Abbreviation         string     `json:"abbreviation"`
	CreationDate         string     `json:"creationDate"`
	Identifier           string     `json:"identifier"`
	ImportIdentifier     any        `json:"importIdentifier"`
	Name                 string     `json:"name"`
	Provider             any        `json:"provider"`
	ProviderSourceStatus string     `json:"providerSourceStatus"`
	RegionOfAuthority    any        `json:"regionOfAuthority"`
	SubscriptionCategory string     `json:"subscriptionCategory"`
	ModificationDate     any        `json:"modificationDate"`
	Description          any        `json:"description"`
	Type                 SourceType `json:"type"`
}

type SourceType struct {
	Category   SourceCategory `json:"category"`
	Identifier string         `json:"identifier"`
	Name       string         `json:"name"`
}

type SourceCategory struct {
	Description string `json:"description"`
	Identifier  string `json:"identifier"`
	Name        string `json:"name"`
}

type UpdatedDates struct {
	AgeUpdated                 any `json:"ageUpdated"`
	AliasesUpdated             any `json:"aliasesUpdated"`
	AlternativeSpellingUpdated any `json:"alternativeSpellingUpdated"`
	AsOfDateUpdated            any `json:"asOfDateUpdated"`
	CategoryUpdated            any `json:"categoryUpdated"`
	CitizenshipsUpdated        any `json:"citizenshipsUpdated"`
	CompaniesUpdated           any `json:"companiesUpdated"`
	DeceasedUpdated            any `json:"deceasedUpdated"`
	DobsUpdated                any `json:"dobsUpdated"`
	EiUpdated                  any `json:"eiUpdated"`
	EnteredUpdated             any `json:"enteredUpdated"`
	ExternalSourcesUpdated     any `json:"externalSourcesUpdated"`
	FirstNameUpdated           any `json:"firstNameUpdated"`
	ForeignAliasUpdated        any `json:"foreignAliasUpdated"`
	FurtherInformationUpdated  any `json:"furtherInformationUpdated"`
	IdNumbersUpdated           any `json:"idNumbersUpdated"`
	KeywordsUpdated            any `json:"keywordsUpdated"`
	LastNameUpdated            any `json:"lastNameUpdated"`
	LinkedToUpdated            any `json:"linkedToUpdated"`
	LocationsUpdated           any `json:"locationsUpdated"`
	LowQualityAliasesUpdated   any `json:"lowQualityAliasesUpdated"`
	PassportsUpdated           any `json:"passportsUpdated"`
	PlaceOfBirthUpdated        any `json:"placeOfBirthUpdated"`
	PositionUpdated            any `json:"positionUpdated"`
	SsnUpdated                 any `json:"ssnUpdated"`
	SubCategoryUpdated         any `json:"subCategoryUpdated"`
	TitleUpdated               any `json:"titleUpdated"`
	UpdatecategoryUpdated      any `json:"updatecategoryUpdated"`
	SicUpdated                 any `json:"sicUpdated"`
}

type Weblink struct {
	Caption any      `json:"caption"`
	Uri     string   `json:"uri"`
	Tags    []string `json:"tags"`
}

type Event struct {
	Address          EventAddress `json:"address"`
	AllegedAddresses []any        `json:"allegedAddresses"`
	Day              int          `json:"day"`
	FullDate         string       `json:"fullDate"`
	Month            int          `json:"month"`
	Type             string       `json:"type"`
	Year             int          `json:"year"`
}

type EventAddress struct {
	City     any     `json:"city"`
	Country  Country `json:"country"`
	PostCode any     `json:"postCode"`
	Region   string  `json:"region"`
	Street   any     `json:"street"`
}

type Role struct {
	End      any    `json:"end"`
	Location any    `json:"location"`
	Source   any    `json:"source"`
	Start    any    `json:"start"`
	Title    string `json:"title"`
	Type     string `json:"type"`
}
