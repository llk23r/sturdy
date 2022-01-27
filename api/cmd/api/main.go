package main

import (
	"context"
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	module_analytics "getsturdy.com/api/pkg/analytics/module"
	"getsturdy.com/api/pkg/api"
	module_api "getsturdy.com/api/pkg/api/module"
	module_auth "getsturdy.com/api/pkg/auth/module"
	module_author "getsturdy.com/api/pkg/author/module"
	module_change "getsturdy.com/api/pkg/change/module"
	service_change "getsturdy.com/api/pkg/change/service"
	module_ci "getsturdy.com/api/pkg/ci/module"
	service_ci "getsturdy.com/api/pkg/ci/service"
	module_codebase_acl "getsturdy.com/api/pkg/codebase/acl/module"
	module_codebase "getsturdy.com/api/pkg/codebase/module"
	module_comments "getsturdy.com/api/pkg/comments/module"
	"getsturdy.com/api/pkg/db"
	"getsturdy.com/api/pkg/di"
	"getsturdy.com/api/pkg/emails"
	module_transactional "getsturdy.com/api/pkg/emails/transactional/module"
	module_features "getsturdy.com/api/pkg/features/module"
	module_file "getsturdy.com/api/pkg/file/module"
	module_gc "getsturdy.com/api/pkg/gc/module"
	"getsturdy.com/api/pkg/github/config"
	module_github "getsturdy.com/api/pkg/github/module"
	module_gitserver "getsturdy.com/api/pkg/gitserver"
	module_graphql "getsturdy.com/api/pkg/graphql"
	"getsturdy.com/api/pkg/http"
	module_http "getsturdy.com/api/pkg/http/module"
	module_installations "getsturdy.com/api/pkg/installations/module"
	module_integrations "getsturdy.com/api/pkg/integrations/module"
	module_jwt "getsturdy.com/api/pkg/jwt/module"
	module_license "getsturdy.com/api/pkg/licenses/module"
	"getsturdy.com/api/pkg/metrics/zapprometheus"
	module_mutagen "getsturdy.com/api/pkg/mutagen/module"
	module_newsletter "getsturdy.com/api/pkg/newsletter/module"
	module_notification "getsturdy.com/api/pkg/notification/module"
	module_onboarding "getsturdy.com/api/pkg/onboarding/module"
	module_onetime "getsturdy.com/api/pkg/onetime/module"
	module_organization "getsturdy.com/api/pkg/organization/module"
	module_pki "getsturdy.com/api/pkg/pki/module"
	module_presence "getsturdy.com/api/pkg/presence/module"
	"getsturdy.com/api/pkg/queue"
	module_review "getsturdy.com/api/pkg/review/module"
	module_servicetokens "getsturdy.com/api/pkg/servicetokens/module"
	module_snapshots "getsturdy.com/api/pkg/snapshots/module"
	db_statuses "getsturdy.com/api/pkg/statuses/db"
	module_statuses "getsturdy.com/api/pkg/statuses/module"
	service_statuses "getsturdy.com/api/pkg/statuses/service"
	module_suggestions "getsturdy.com/api/pkg/suggestions/module"
	service_sync "getsturdy.com/api/pkg/sync/service"
	db_user "getsturdy.com/api/pkg/user/db"
	module_user "getsturdy.com/api/pkg/user/module"
	db_view "getsturdy.com/api/pkg/view/db"
	"getsturdy.com/api/pkg/view/events"
	meta_view "getsturdy.com/api/pkg/view/meta"
	"getsturdy.com/api/pkg/view/view_workspace_snapshot"
	"getsturdy.com/api/pkg/waitinglist"
	"getsturdy.com/api/pkg/waitinglist/acl"
	"getsturdy.com/api/pkg/waitinglist/instantintegration"
	db_activity "getsturdy.com/api/pkg/workspace/activity/db"
	activity_sender "getsturdy.com/api/pkg/workspace/activity/sender"
	service_activity "getsturdy.com/api/pkg/workspace/activity/service"
	db_workspace "getsturdy.com/api/pkg/workspace/db"
	ws_meta "getsturdy.com/api/pkg/workspace/meta"
	module_workspace "getsturdy.com/api/pkg/workspace/module"
	db_workspace_watchers "getsturdy.com/api/pkg/workspace/watchers/db"
	service_workspace_watchers "getsturdy.com/api/pkg/workspace/watchers/service"
	"getsturdy.com/api/vcs/executor"
	"getsturdy.com/api/vcs/provider"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func main() {
	reposBasePath := flag.String("repos-base-path", "tmp/repos", "path on the filesystem to where all repos can be found")
	httpListenAddr := flag.String("http-listen-addr", "127.0.0.1:3000", "")
	httpPprofListenAddr := flag.String("http-pprof-listen-addr", "127.0.0.1:6060", "")
	gitListenAddr := flag.String("git-listen-addr", "127.0.0.1:3002", "")
	metricsListenAddr := flag.String("metrics-listen-addr", "127.0.0.1:2112", "")
	dbSourceAddr := flag.String("db", "postgres://mash:mash@127.0.0.1:5432/mash?sslmode=disable", "")
	productionLogger := flag.Bool("production-logger", false, "")
	gitHubAppID := flag.Int64("github-app-id", 122610, "")
	gitHubAppName := flag.String("github-app-name", "sturdy-gustav-localhost", "")
	gitHubAppClientID := flag.String("github-app-client-id", "", "")
	gitHubAppSecret := flag.String("github-app-secret", "", "")
	gitHubAppPrivateKeyPath := flag.String("github-app-private-key-path", "", "")
	gitLfsHostname := flag.String("git-lfs-hostname", "localhost:8888", "")
	enableTransactionalEmails := flag.Bool("enable-transactional-emails", false, "")
	exportBucketName := flag.String("export-bucket-name", "", "the S3 bucket to be used for change exports")
	developmentAllowExtraCorsOrigin := flag.String("development-allow-extra-cors-origin", "", "Additional CORS origin to be allowed")
	localQueue := flag.Bool("use-local-queues", false, "If set, local queue will be used instead of SQS")

	// deprecated flags
	_ = flag.Bool("send-posthog-events", false, "")
	_ = flag.Bool("send-invites-worker", false, "")
	_ = flag.String("gmail-token-json-path", "", "used by the invite email sender")
	_ = flag.Bool("unauthenticated-graphql-introspection", false, "")
	_ = flag.String("gmail-credentials-json-path", "", "used by the invite email sender")

	publicApiHostname := flag.String("public-api-hostname", "localhost", "api.getsturdy.com in production")
	// publicGitHostname := flag.String("public-git-hostname", "git.getsturdy.com", "")

	defaultHostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}
	hostname := flag.String("hostname", defaultHostname, "")
	queuePrefix := flag.String("queue-prefix", "dev"+defaultHostname, "set to 'production' when running in production")

	flag.Parse()

	ctx := context.Background()

	var logger *zap.Logger
	if *productionLogger {
		logger, _ = zap.NewProduction(zap.Hooks(zapprometheus.Hook))
	} else {
		logger, _ = zap.NewDevelopment(zap.Hooks(zapprometheus.Hook))
	}

	providers := []interface{}{
		func() context.Context {
			return ctx
		},
		func() *zap.Logger { return logger },
		func() provider.RepoProvider {
			return provider.New(*reposBasePath, *gitLfsHostname)
		},
		func() (*sqlx.DB, error) {
			return db.TrySetup(logger, *dbSourceAddr, 5*time.Second)
		},
		func() config.GitHubAppConfig {
			return config.GitHubAppConfig{
				GitHubAppID:             *gitHubAppID,
				GitHubAppName:           *gitHubAppName,
				GitHubAppClientID:       *gitHubAppClientID,
				GitHubAppSecret:         *gitHubAppSecret,
				GitHubAppPrivateKeyPath: *gitHubAppPrivateKeyPath,
			}
		},
		func() (*session.Session, error) {
			awsSession, err := session.NewSession(
				&aws.Config{
					Region: aws.String("eu-north-1"),
				})
			return awsSession, err
		},
		func(e events.EventReadWriter) events.EventReader {
			return e
		},
		func(awsSession *session.Session) (queue.Queue, error) {
			if *localQueue {
				return queue.NewInMemory(logger), nil
			} else {
				return queue.NewSQS(logger, awsSession, *hostname, *queuePrefix)
			}
		},
		func(awsSession *session.Session) emails.Sender {
			if *enableTransactionalEmails {
				return emails.NewSES(awsSession)
			}
			return emails.NewLogs(logger)
		},
		func() service_ci.PublicAPIHostname {
			return service_ci.PublicAPIHostname(*publicApiHostname)
		},
		func() service_change.ExportBucketName {
			return service_change.ExportBucketName(exportBucketName)
		},
		func(repo db_workspace.Repository) db_workspace.WorkspaceReader {
			return repo
		},
		func() http.DevelopmentAllowExtraCorsOrigin {
			return http.DevelopmentAllowExtraCorsOrigin(*developmentAllowExtraCorsOrigin)
		},
		events.NewInMemory,
		executor.NewProvider,
		events.NewSender,
		service_activity.New,
		activity_sender.NewActivitySender,
		ws_meta.NewWriterWithEvents,
		meta_view.NewViewUpdatedFunc,
		service_statuses.New,
		service_workspace_watchers.New,
		db_user.NewRepo,
		db_view.NewRepo,
		db_workspace.NewRepo,
		waitinglist.NewWaitingListRepo,
		acl.NewACLInterestRepo,
		instantintegration.NewInstantIntegrationInterestRepo,
		view_workspace_snapshot.NewRepo,
		db_activity.NewActivityRepo,
		db_activity.NewActivityReadsRepo,
		db_statuses.New,
		db_workspace_watchers.NewDB,
		service_sync.New,
	}

	mainModule := func(c *di.Container) {
		for _, provider := range providers {
			c.Register(provider)
		}

		c.Import(module_analytics.Module)
		c.Import(module_api.Module)
		c.Import(module_auth.Module)
		c.Import(module_author.Module)
		c.Import(module_change.Module)
		c.Import(module_ci.Module)
		c.Import(module_codebase.Module)
		c.Import(module_codebase_acl.Module)
		c.Import(module_comments.Module)
		c.Import(module_features.Module)
		c.Import(module_file.Module)
		c.Import(module_gc.Module)
		c.Import(module_github.Module)
		c.Import(module_gitserver.Module)
		c.Import(module_graphql.Module)
		c.Import(module_http.Module)
		c.Import(module_installations.Module)
		c.Import(module_integrations.Module)
		c.Import(module_jwt.Module)
		c.Import(module_license.Module)
		c.Import(module_mutagen.Module)
		c.Import(module_newsletter.Module)
		c.Import(module_notification.Module)
		c.Import(module_onboarding.Module)
		c.Import(module_onetime.Module)
		c.Import(module_organization.Module)
		c.Import(module_pki.Module)
		c.Import(module_presence.Module)
		c.Import(module_review.Module)
		c.Import(module_servicetokens.Module)
		c.Import(module_snapshots.Module)
		c.Import(module_statuses.Module)
		c.Import(module_suggestions.Module)
		c.Import(module_user.Module)

		// todo: continue importing here

		c.Import(module_transactional.Module)
		c.Import(module_workspace.Module)
	}

	var apiServer api.Starter
	if err := di.Init(&apiServer, mainModule); err != nil {
		log.Fatalf("%+v", err)
	}

	if err := apiServer.Start(ctx, &api.Config{
		GitListenAddr:       *gitListenAddr,
		HTTPPProfListenAddr: *httpPprofListenAddr,
		MetricsListenAddr:   *metricsListenAddr,
		HTTPAddr:            *httpListenAddr,
	}); err != nil {
		logger.Fatal("failed to start api server", zap.Error(err))
	}
}