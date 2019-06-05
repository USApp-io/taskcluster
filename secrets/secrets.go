package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/taskcluster/httpbackoff"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/tcsecrets"
	"github.com/taskcluster/taskcluster-worker-runner/cfg"
	"github.com/taskcluster/taskcluster-worker-runner/runner"
	"github.com/taskcluster/taskcluster-worker-runner/tc"
)

func clientFactory(rootURL string, credentials *tcclient.Credentials) (tc.Secrets, error) {
	return tcsecrets.New(credentials, rootURL), nil
}

func ConfigureRun(cfg *cfg.RunnerConfig, run *runner.Run) error {
	return configureRun(cfg, run, clientFactory)
}

func configureRun(runnercfg *cfg.RunnerConfig, run *runner.Run, secretsClientFactory tc.SecretsClientFactory) error {
	secretsClient, err := secretsClientFactory(run.RootURL, &run.Credentials)
	if err != nil {
		return err
	}

	secretName := "worker-type:" + run.WorkerPoolID
	secResponse, err := secretsClient.Get(secretName)
	if err != nil {
		// 404 error is ok, since secrets aren't required. Anything else indicates there was a problem retrieving
		// secret or talking to secrets service, so they should return an error
		if apiCallException, isAPICallException := err.(*tcclient.APICallException); isAPICallException {
			rootCause := apiCallException.RootCause
			if badHTTPResponseCode, isBadHTTPResponseCode := rootCause.(httpbackoff.BadHttpResponseCode); isBadHTTPResponseCode {
				if badHTTPResponseCode.HttpResponseCode == 404 {
					log.Printf("WARNING: No worker secrets for worker pool %v - secret %v does not exist.", run.WorkerPoolID, secretName)
					return nil
				}
			}
		}
		return err
	}

	// some secrets contain raw configuration, while others contain the preferred {config: .., files: ..}.  If we have
	// something of the latter shape, we assume that's what we've got, and otherwise make itup
	var secret struct {
		Config *cfg.WorkerConfig `yaml:"config"`
		Files  []interface{}     `yaml:"files"`
	}

	decoder := json.NewDecoder(bytes.NewReader(secResponse.Secret))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&secret)
	if err != nil {
		log.Printf("Falling back to legacy secret format without top-level config/files properties")
		err := json.Unmarshal(secResponse.Secret, &secret.Config)
		fmt.Printf("%#v\n", secret.Config)
		if err != nil {
			return fmt.Errorf("Secret value is not a JSON object")
		}
	}

	run.WorkerConfig = run.WorkerConfig.Merge(secret.Config)
	fmt.Printf("%#v\n", run.WorkerConfig)

	if len(secret.Files) != 0 {
		return fmt.Errorf("secret files are nonempty - files are not supported yet")
	}

	return nil
}
