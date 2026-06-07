package btcx

import (
	"context"
	"testing"
)

type fakeRPCClient struct {
	listUnspentResult []ListUnspentResult
	listUnspentErr    error
	createResult      string
	createErr         error
	signResult        *SignRawTransactionWithWalletResult
	signErr           error
	sendResult        string
	sendErr           error
}

type fakeAddressRPCClient struct {
	getNewAddressFn func(ctx context.Context, label string, addressType string) (string, error)
}

func (f *fakeRPCClient) ListUnspent(context.Context, int, int, []string) ([]ListUnspentResult, error) {
	return f.listUnspentResult, f.listUnspentErr
}

func (f *fakeRPCClient) CreateRawTransaction(context.Context, []Input, []map[string]any) (string, error) {
	return f.createResult, f.createErr
}

func (f *fakeRPCClient) SignRawTransactionWithWallet(context.Context, string) (*SignRawTransactionWithWalletResult, error) {
	return f.signResult, f.signErr
}

func (f *fakeRPCClient) SendRawTransaction(context.Context, string) (string, error) {
	return f.sendResult, f.sendErr
}

func (f *fakeAddressRPCClient) GetNewAddress(ctx context.Context, label string, addressType string) (string, error) {
	return f.getNewAddressFn(ctx, label, addressType)
}

func TestNewWithdrawSenderValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	if _, err := NewWithdrawSender(NodeConfig{
		Username: "user",
		Password: "pass",
	}); err == nil {
		t.Fatal("NewWithdrawSender() should fail when url is missing")
	}
	if _, err := NewWithdrawSender(NodeConfig{
		URL:      "http://127.0.0.1:18332",
		Password: "pass",
	}); err == nil {
		t.Fatal("NewWithdrawSender() should fail when username is missing")
	}
	if _, err := NewWithdrawSender(NodeConfig{
		URL:      "http://127.0.0.1:18332",
		Username: "user",
	}); err == nil {
		t.Fatal("NewWithdrawSender() should fail when password is missing")
	}
}

func TestNewAddressAllocatorValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	if _, err := NewAddressAllocator(NodeConfig{
		Username: "user",
		Password: "pass",
	}); err == nil {
		t.Fatal("NewAddressAllocator() should fail when url is missing")
	}
	if _, err := NewAddressAllocator(NodeConfig{
		URL:      "http://127.0.0.1:18332/wallet/mscoin",
		Password: "pass",
	}); err == nil {
		t.Fatal("NewAddressAllocator() should fail when username is missing")
	}
	if _, err := NewAddressAllocator(NodeConfig{
		URL:      "http://127.0.0.1:18332/wallet/mscoin",
		Username: "user",
	}); err == nil {
		t.Fatal("NewAddressAllocator() should fail when password is missing")
	}
}

func TestSenderSendBuildsAndBroadcastsTransaction(t *testing.T) {
	t.Parallel()

	service := &sender{
		client: &fakeRPCClient{
			listUnspentResult: []ListUnspentResult{
				{Txid: "tx-1", Vout: 0, Amount: 1.2},
			},
			createResult: "raw-hex",
			signResult: &SignRawTransactionWithWalletResult{
				Hex:      "signed-hex",
				Complete: true,
			},
			sendResult: "final-txid",
		},
		minConfirmations: 0,
		maxConfirmations: 999999,
	}

	txid, err := service.Send(context.Background(), "source", "target", 1.0, 0.95)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if txid != "final-txid" {
		t.Fatalf("Send() txid = %q, want final-txid", txid)
	}
}

func TestSenderSendRejectsInsufficientBalance(t *testing.T) {
	t.Parallel()

	service := &sender{
		client: &fakeRPCClient{
			listUnspentResult: []ListUnspentResult{
				{Txid: "tx-1", Vout: 0, Amount: 0.2},
			},
		},
		minConfirmations: 0,
		maxConfirmations: 999999,
	}

	if _, err := service.Send(context.Background(), "source", "target", 1.0, 0.95); err == nil {
		t.Fatal("Send() should fail when utxo balance is insufficient")
	}
}

func TestAddressAllocatorAllocateUsesConfiguredAddressType(t *testing.T) {
	t.Parallel()

	allocator := &addressAllocator{
		client: &fakeAddressRPCClient{
			getNewAddressFn: func(ctx context.Context, label string, addressType string) (string, error) {
				if label != "member-8-btc" {
					t.Fatalf("GetNewAddress() label = %q, want member-8-btc", label)
				}
				if addressType != "legacy" {
					t.Fatalf("GetNewAddress() addressType = %q, want legacy", addressType)
				}
				return "mtestAllocatedAddress", nil
			},
		},
		addressType: "legacy",
	}

	address, err := allocator.Allocate(context.Background(), "member-8-btc")
	if err != nil {
		t.Fatalf("Allocate() error = %v", err)
	}
	if address != "mtestAllocatedAddress" {
		t.Fatalf("Allocate() = %q, want mtestAllocatedAddress", address)
	}
}
