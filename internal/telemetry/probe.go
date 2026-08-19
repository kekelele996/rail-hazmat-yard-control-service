package telemetry

func DecoderAvailable(d Decoder) bool { return !isNilDecoder(d) }
