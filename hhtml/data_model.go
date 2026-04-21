package hhtml

// DataAttributeInfo represents the metadata for an attribute.
type DataAttributeInfo struct {
	Deprecated   bool `json:"deprecated,omitempty"`
	Experimental bool `json:"experimental,omitempty"`
}

// DataTag represents the metadata for a tag, including its attributes.
type DataTag struct {
	Deprecated       bool                         `json:"deprecated,omitempty"`
	Experimental     bool                         `json:"experimental,omitempty"`
	Attributes       map[string]DataAttributeInfo `json:"attributes,omitempty"`
	DescriptionShort string                       `json:"description_short,omitempty"`
	DescriptionLong  string                       `json:"description_long,omitempty"`
}

// TagData is the top-level structure for all tags.
type TagData map[string]DataTag

// You can load the JSON into this structure using encoding/json.
// Example:
//   var tags TagData
//   err := json.Unmarshal(data, &tags)
