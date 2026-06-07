package btcx

import "testing"

func TestWalletGeneratesTestnetAddressAndPrivateKey(t *testing.T) {
	wallet, err := NewWallet()
	if err != nil {
		t.Fatalf("NewWallet() error = %v", err)
	}

	address := wallet.TestnetAddress()
	if address == "" {
		t.Fatal("TestnetAddress() returned empty address")
	}

	key, err := wallet.EncodedPrivateKey()
	if err != nil {
		t.Fatalf("EncodedPrivateKey() error = %v", err)
	}
	if key == "" {
		t.Fatal("EncodedPrivateKey() returned empty key")
	}
}
