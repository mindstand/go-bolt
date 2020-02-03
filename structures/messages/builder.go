package messages

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"time"
)

const (
	bookmarksMetadataKey  = "bookmarks"
	databaseNameKey       = "db"
	txTimeoutMetadataKey  = "tx_timeout"
	txMetadataMetadataKey = "tx_metadata"
	modeKey               = "mode"
	modeReadValue         = "r"
	// todo update with db name
	defaultDbName = ""
)

func BuildTxMetadata(txTimeout *time.Duration, txMetadata map[string]interface{}, mode bolt_mode.AccessMode, bookmark interface{}) map[string]interface{} {
	return BuildTxMetadataWithDatabase(txTimeout, txMetadata, defaultDbName, mode, bookmark)
}

func BuildTxMetadataWithDatabase(txTimeout *time.Duration, txMetadata map[string]interface{}, databaseName string, mode bolt_mode.AccessMode, bookmark interface{}) map[string]interface{} {
	// todo replace once bookmarks are supported
	bookmarksPresent := false
	txTimeoutPresent := txTimeout != nil && *txTimeout != 0
	txMetaDataPresent := txMetadata != nil && len(txMetadata) != 0
	accessModePresent := mode != bolt_mode.ReadMode
	databaseNamePresent := databaseName != ""

	if !bookmarksPresent && !txTimeoutPresent && !txMetaDataPresent && !accessModePresent && !databaseNamePresent {
		return map[string]interface{}{}
	}

	toReturn := map[string]interface{}{}

	if bookmarksPresent {
		// todo support bookmarking
	}

	if txTimeoutPresent {
		toReturn[txTimeoutMetadataKey] = txTimeout.Milliseconds()
	}

	if txMetaDataPresent {
		toReturn[txMetadataMetadataKey] = txMetadata
	}

	if accessModePresent {
		toReturn[modeKey] = modeReadValue
	}

	if databaseNamePresent {
		toReturn[databaseNameKey] = databaseName
	}

	return toReturn
}
