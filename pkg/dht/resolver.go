package dht

// The ContentResolver interface is used to insert and query content.
type ContentResolver interface {
	// Insert content with a specific content ID. Usually, the content ID will
	// stores information about the type and the hash of the content.
	InsertContent(contentID, content []byte)

	// QueryContent returns the content associated with a content ID. If there
	// is no associated content, it returns false. Otherwise, it returns true.
	QueryContent(contentID []byte) (content []byte, contentOk bool)
}
