package gssapi

import (
	"context"
	"encoding/asn1"
	"encoding/binary"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/crypto"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/iana/chksumtype"
	"github.com/jcmturner/gokrb5/v8/iana/keyusage"
	"github.com/jcmturner/gokrb5/v8/messages"
	"github.com/jcmturner/gokrb5/v8/types"
	"github.com/segmentio/kafka-go/sasl"
)

const (
	// TOK_ID_KRB_AP_REQ see: https://tools.ietf.org/html/rfc4121#section-4.1
	TOK_ID_KRB_AP_REQ = "\x01\x00"
)

type mechanism struct {
	client      *client.Client
	serviceName string
}

func (m mechanism) Name() string {
	return "GSSAPI"
}

// GoKRB5v8 uses gokrb5/v8 to implement the GSSAPI mechanism.
//
// client is a github.com/gokrb5/v8/client *Client instance.
// kafkaServiceName is the name of the Kafka service in your Kerberos.
func GoKRB5v8(client *client.Client, kafkaServiceName string) sasl.Mechanism {
	return mechanism{client: client, serviceName: kafkaServiceName}
}

func (m mechanism) Start(ctx context.Context) (sasl.StateMachine, []byte, error) {
	// Get metadata from context
	metadata := sasl.MetadataFromContext(ctx)

	// Setup service principle name
	servicePrincipalName := m.serviceName + "/" + metadata.Host

	// Get the service ticket
	ticket, key, err := m.client.GetServiceTicket(
		servicePrincipalName,
	)
	if err != nil {
		return nil, nil, err
	}

	// Set up a new authenticator
	authenticator, err := types.NewAuthenticator(
		m.client.Credentials.Realm(),
		m.client.Credentials.CName(),
	)
	if err != nil {
		return nil, nil, err
	}

	// Get the encryption type from the key
	encryptionType, err := crypto.GetEtype(key.KeyType)
	if err != nil {
		return nil, nil, err
	}

	// Get the keysize and generate sequence number and sub key
	keySize := encryptionType.GetKeyByteSize()
	err = authenticator.GenerateSeqNumberAndSubKey(key.KeyType, keySize)
	if err != nil {
		return nil, nil, err
	}

	// Set the checksum type
	authenticator.Cksum = types.Checksum{
		CksumType: chksumtype.GSSAPI,
		Checksum:  authenticatorPseudoChecksum(),
	}

	// Set up new KRB_AP_REQ
	apReq, err := messages.NewAPReq(ticket, key, authenticator)
	if err != nil {
		return nil, nil, err
	}

	// Marshal the KRB_AP_REQ
	bytes, err := apReq.Marshal()
	if err != nil {
		return nil, nil, err
	}

	bytesWithPrefix := make([]byte, 0, len(TOK_ID_KRB_AP_REQ)+len(bytes))
	bytesWithPrefix = append(bytesWithPrefix, TOK_ID_KRB_AP_REQ...)
	bytesWithPrefix = append(bytesWithPrefix, bytes...)

	// Get the token
	gssapiToken, err := prependGSSAPITokenTag(bytesWithPrefix)
	if err != nil {
		return nil, nil, err
	}

	return &gokrb5v8Session{authenticator.SubKey, false}, gssapiToken, nil
}

func authenticatorPseudoChecksum() []byte {
	// Not actually a checksum, but it goes in the checksum field.
	// https://tools.ietf.org/html/rfc4121#section-4.1.1
	checksum := make([]byte, 24)

	flags := gssapi.ContextFlagInteg
	// Reasons for each flag being on or off:
	//     Delegation: Off. We are not using delegated credentials.
	//     Mutual: Off. Mutual authentication is already provided
	//         as a result of how Kerberos works.
	//     Replay: Off. We don’t need replay protection because each
	//         packet is secured by a per-session key and is unique
	//         within its session.
	//     Sequence: Off. Out-of-order messages cannot happen in our
	//         case, and if it somehow happened anyway it would
	//         necessarily trigger other appropriate errors.
	//     Confidentiality: Off. Our authentication itself does not
	//         seem to be requesting or using any “security layers”
	//         in the GSSAPI sense, and this is just one of the
	//         security layer features. Also, if we were requesting
	//         a GSSAPI security layer, we would be required to
	//         set the mutual flag to on.
	//         https://tools.ietf.org/html/rfc4752#section-3.1
	//     Integrity: On. Must be on when calling the standard API,
	//         so it probably must be set in the raw packet itself.
	//         https://tools.ietf.org/html/rfc4752#section-3.1
	//         https://tools.ietf.org/html/rfc4752#section-7
	//     Anonymous: Off. We are not using an anonymous ticket.
	//         https://tools.ietf.org/html/rfc6112#section-3

	binary.LittleEndian.PutUint32(checksum[0:4], 16)
	// checksum[4:20] is unused/blank channel binding settings.
	binary.LittleEndian.PutUint32(checksum[20:24], uint32(flags))
	return checksum
}

type gssapiToken struct {
	OID    asn1.ObjectIdentifier
	Object asn1.RawValue
}

func prependGSSAPITokenTag(payload []byte) ([]byte, error) {
	// The GSSAPI "token" is almost an ASN.1 encoded object, except
	// that the "token object" is raw bytes, not necessarily ASN.1.
	// https://tools.ietf.org/html/rfc2743#page-81 (section 3.1)
	token := gssapiToken{
		OID:    asn1.ObjectIdentifier(gssapi.OIDKRB5.OID()),
		Object: asn1.RawValue{FullBytes: payload},
	}
	return asn1.MarshalWithParams(token, "application")
}

type gokrb5v8Session struct {
	key  types.EncryptionKey
	done bool
}

func (s *gokrb5v8Session) Next(ctx context.Context, challenge []byte) (bool, []byte, error) {
	if s.done {
		return true, nil, nil
	}
	const tokenIsFromGSSAcceptor = true
	challengeToken := gssapi.WrapToken{}
	err := challengeToken.Unmarshal(challenge, tokenIsFromGSSAcceptor)
	if err != nil {
		return false, nil, err
	}

	valid, err := challengeToken.Verify(
		s.key,
		keyusage.GSSAPI_ACCEPTOR_SEAL,
	)
	if !valid {
		return false, nil, err
	}

	responseToken, err := gssapi.NewInitiatorWrapToken(
		challengeToken.Payload,
		s.key,
	)
	if err != nil {
		return false, nil, err
	}

	response, err := responseToken.Marshal()
	if err != nil {
		return false, nil, err
	}

	// We are done, but we can't return `true` yet because
	// the SASL loop calling this needs the first return to be
	// `false` any time there are response bytes to send.
	s.done = true
	return false, response, nil
}
