package handlers

const (
	InvalidRequest                    = "Invalid Request."
	InsertionError                    = "Error during insertion."
	InvalidContentEncryptionAlgorithm = "Invalid Content Encryption Algorithm. Must be AES 256 GCM"
	InvalidKeyEncryptionAlgorithm     = "Invalid Key Encryption Algorithm. Must be RSA OEAP with SHA-256"
	InvalidAmountOfRecipients         = "The message contains too many recipients. Expected just one."
	StoreMessageFailed                = "Message couldnt be stored."
	NoBodyError                       = "No Body."
	BodyParseError                    = "Body could ne parsed."
	WrongContentType                  = "Wrong Content Type."
	DeviceAlreadyExist                = "Device already exist."
	DeviceRegistrationFailed          = "Device Registration failed."
	InvalidKeySigningAlgorithm        = "Invalid Key Signing Algorithm"
	CryptoProviderError               = "Error happened in Crypto Provider"
)
