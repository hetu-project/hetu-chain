package types

// GetSignatureByAddress returns the BLS signature for a given Ethereum address
func (m *MsgBLSCallback) GetSignatureByAddress(address string) (string, bool) {
	for _, addrSig := range m.AddrSigs {
		if addrSig.Address == address {
			return addrSig.Signature, true
		}
	}
	return "", false
}
