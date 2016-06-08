package firebase

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
)

const (
	clientCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

var (
	googleProjectID = os.Getenv("GOOGLE_PROJECT_ID")
)

/*
FirebaseTokenGenerator.prototype.verifyIdToken = function(idToken) {
  if (typeof idToken !== 'string') {
    throw new Error('First argument to verifyIdToken() must be a Firebase Auth ID token');
  }

  if (typeof this.serviceAccount.project_id !== 'string' || this.serviceAccount.project_id === '') {
    throw new Error('verifyIdToken() requires a service account with project_id set');
  }

  var fullDecodedToken = jwt.decode(idToken, {
    complete: true
  });

  var header = fullDecodedToken && fullDecodedToken.header;
  var payload = fullDecodedToken && fullDecodedToken.payload;

  var errorMessage;
  if (!fullDecodedToken) {
    errorMessage = 'Decoding Firebase Auth ID token failed';
  } else if (typeof header.kid === 'undefined') {
    errorMessage = 'Firebase Auth ID token has no "kid" claim';
  } else if (header.alg !== ALGORITHM) {
    errorMessage = 'Firebase Auth ID token has incorrect algorithm';
  } else if (payload.aud !== this.serviceAccount.project_id) {
    errorMessage = 'Firebase Auth ID token has incorrect "aud" claim';
  } else if (payload.iss !== 'https://securetoken.google.com/' + this.serviceAccount.project_id) {
    errorMessage = 'Firebase Auth ID token has incorrect "iss" claim';
  } else if (typeof payload.sub !== 'string' || payload.sub === '' || payload.sub.length > 128) {
    errorMessage = 'Firebase Auth ID token has invalid "sub" claim';
  }

  if (typeof errorMessage !== 'undefined') {
    return firebase.Promise.reject(new Error(errorMessage));
  }

  return this._fetchPublicKeys().then(function(publicKeys) {
    if (!publicKeys.hasOwnProperty(header.kid)) {
      return firebase.Promise.reject('Firebase Auth ID token has "kid" claim which does not correspond to a known public key');
    }

    return new firebase.Promise(function(resolve, reject) {
      jwt.verify(idToken, publicKeys[header.kid], {
        algorithms: [ALGORITHM]
      }, function(error, decodedToken) {
        if (error) {
          reject(error);
        } else {
          resolve(decodedToken);
        }
      });
    });
  });
};
*/

func VerifyIDToken(idToken string) (string, error) {
	keys, err := fetchPublicKeys()

	if err != nil {
		return "", err
	}

	parsedToken, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		kid := token.Claims["kid"]

		certPEM := []byte(*keys[kid])
		block, _ := pem.Decode([]byte(certPEM))
		var cert *x509.Certificate
		cert, _ = x509.ParseCertificate(block.Bytes)
		rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)

		return rsaPublicKey, nil
	})

	if err != nil {
		return "", err
	}

	var errMessage *string

	if parsedToken.Claims["aud"] != googleProjectID {
		*errMessage = "Firebase Auth ID token has incorrect 'aud' claim"
	} else if parsedToken.Claims["iss"] != "https://securetoken.google.com/"+googleProjectID {
		*errMessage = "Firebase Auth ID token has incorrect 'iss' claim"
	} else if parsedToken.Claims["sub"] == "" || len(string(token.Claims["sub"].(string))) > 128 {
		*errMessage = "Firebase Auth ID token has invalid 'sub' claim"
	}

	if errMessage != nil {
		return "", errors.New(*errMessage)
	}

	return string(parsedToken.Claims["sub"].(string)), nil
}

func fetchPublicKeys() (map[string]*json.RawMessage, error) {
	resp, err := http.Get(clientCertURL)

	if err != nil {
		return nil, err
	}

	var objmap map[string]*json.RawMessage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&objmap)

	return objmap, err
}
