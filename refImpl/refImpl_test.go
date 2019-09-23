package refImpl

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-ocf/kit/security"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	var config Config
	os.Setenv("OAUTH_CALLBACK", "OAUTH_CALLBACK")
	os.Setenv("EVENTS_URL", "EVENTS_URL")
	os.Setenv("NAME", "NAME")
	os.Setenv("CLIENT_ID", "CLIENT_ID")
	os.Setenv("CLIENT_SECRET", "CLIENT_SECRET")
	os.Setenv("SCOPES", "SCOPES")
	os.Setenv("AUTH_URL", "AUTH_URL")
	os.Setenv("TOKEN_URL", "TOKEN_URL")
	err := envconfig.Process("", &config)
	require.NoError(t, err)
	config.GoRoutinePoolSize = 1
	dir, err := ioutil.TempDir("", "gotesttmp")
	require.NoError(t, err)
	config.Service.TLSConfig = testSetupTLS(t, dir)

	got, err := Init(config)
	require.NoError(t, err)
	require.NotEmpty(t, got)
}

func testSetupTLS(t *testing.T, dir string) security.TLSConfig {
	crt := filepath.Join(dir, "cert.crt")
	if err := ioutil.WriteFile(crt, CertPEMBlock, 0600); err != nil {
		require.NoError(t, err)
	}
	crtKey := filepath.Join(dir, "cert.key")
	if err := ioutil.WriteFile(crtKey, KeyPEMBlock, 0600); err != nil {
		require.NoError(t, err)
	}
	caRootCrt := filepath.Join(dir, "caRoot.crt")
	if err := ioutil.WriteFile(caRootCrt, CARootPemBlock, 0600); err != nil {
		require.NoError(t, err)
	}
	/*
		caInterCrt := filepath.Join(dir, "caInter.crt")
		if err := ioutil.WriteFile(caInterCrt, CAIntermediatePemBlock, 0600); err != nil {
			require.NoError(t, err)
		}
	*/
	return security.TLSConfig{
		Certificate:    crt,
		CertificateKey: crtKey,
		CAPool:         caRootCrt,
	}
}

var (
	CertIdentity = "b5a2a42e-b285-42f1-a36b-034c8fc8efd5"

	CertPEMBlock = []byte(`-----BEGIN CERTIFICATE-----
MIIB9zCCAZygAwIBAgIRAOwIWPAt19w7DswoszkVIEIwCgYIKoZIzj0EAwIwEzER
MA8GA1UEChMIVGVzdCBPUkcwHhcNMTkwNTAyMjAwNjQ4WhcNMjkwMzEwMjAwNjQ4
WjBHMREwDwYDVQQKEwhUZXN0IE9SRzEyMDAGA1UEAxMpdXVpZDpiNWEyYTQyZS1i
Mjg1LTQyZjEtYTM2Yi0wMzRjOGZjOGVmZDUwWTATBgcqhkjOPQIBBggqhkjOPQMB
BwNCAAQS4eiM0HNPROaiAknAOW08mpCKDQmpMUkywdcNKoJv1qnEedBhWne7Z0jq
zSYQbyqyIVGujnI3K7C63NRbQOXQo4GcMIGZMA4GA1UdDwEB/wQEAwIDiDAzBgNV
HSUELDAqBggrBgEFBQcDAQYIKwYBBQUHAwIGCCsGAQUFBwMBBgorBgEEAYLefAEG
MAwGA1UdEwEB/wQCMAAwRAYDVR0RBD0wO4IJbG9jYWxob3N0hwQAAAAAhwR/AAAB
hxAAAAAAAAAAAAAAAAAAAAAAhxAAAAAAAAAAAAAAAAAAAAABMAoGCCqGSM49BAMC
A0kAMEYCIQDuhl6zj6gl2YZbBzh7Th0uu5izdISuU/ESG+vHrEp7xwIhANCA7tSt
aBlce+W76mTIhwMFXQfyF3awWIGjOcfTV8pU
-----END CERTIFICATE-----
`)

	KeyPEMBlock = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMPeADszZajrkEy4YvACwcbR0pSdlKG+m8ALJ6lj/ykdoAoGCCqGSM49
AwEHoUQDQgAEEuHojNBzT0TmogJJwDltPJqQig0JqTFJMsHXDSqCb9apxHnQYVp3
u2dI6s0mEG8qsiFRro5yNyuwutzUW0Dl0A==
-----END EC PRIVATE KEY-----
`)

	CARootPemBlock = []byte(`-----BEGIN CERTIFICATE-----
MIIBaTCCAQ+gAwIBAgIQR33gIB75I7Vi/QnMnmiWvzAKBggqhkjOPQQDAjATMREw
DwYDVQQKEwhUZXN0IE9SRzAeFw0xOTA1MDIyMDA1MTVaFw0yOTAzMTAyMDA1MTVa
MBMxETAPBgNVBAoTCFRlc3QgT1JHMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
xbwMaS8jcuibSYJkCmuVHfeV3xfYVyUq8Iroz7YlXaTayspW3K4hVdwIsy/5U+3U
vM/vdK5wn2+NrWy45vFAJqNFMEMwDgYDVR0PAQH/BAQDAgEGMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wCwYDVR0RBAQwAoIAMAoGCCqGSM49
BAMCA0gAMEUCIBWkxuHKgLSp6OXDJoztPP7/P5VBZiwLbfjTCVRxBvwWAiEAnzNu
6gKPwtKmY0pBxwCo3NNmzNpA6KrEOXE56PkiQYQ=
-----END CERTIFICATE-----		
`)
	CARootKeyPemBlock = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICzfC16AqtSv3wt+qIbrgM8dTqBhHANJhZS5xCpH6P2roAoGCCqGSM49
AwEHoUQDQgAExbwMaS8jcuibSYJkCmuVHfeV3xfYVyUq8Iroz7YlXaTayspW3K4h
VdwIsy/5U+3UvM/vdK5wn2+NrWy45vFAJg==
-----END EC PRIVATE KEY-----	
`)
)
