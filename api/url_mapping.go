package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/kotalco/cloud-api/api/handler/endpoint"
	"github.com/kotalco/cloud-api/api/handler/setting"
	"github.com/kotalco/cloud-api/api/handler/sts"
	"github.com/kotalco/cloud-api/api/handler/subscription"
	"github.com/kotalco/cloud-api/api/handler/svc"
	"github.com/kotalco/cloud-api/api/handler/user"
	"github.com/kotalco/cloud-api/api/handler/workspace"
	"github.com/kotalco/cloud-api/pkg/middleware"
	"github.com/kotalco/community-api/api/handlers/aptos"
	"github.com/kotalco/community-api/api/handlers/bitcoin"
	"github.com/kotalco/community-api/api/handlers/chainlink"
	"github.com/kotalco/community-api/api/handlers/core/secret"
	"github.com/kotalco/community-api/api/handlers/core/storage_class"
	"github.com/kotalco/community-api/api/handlers/ethereum"
	"github.com/kotalco/community-api/api/handlers/ethereum2/beacon_node"
	"github.com/kotalco/community-api/api/handlers/ethereum2/validator"
	"github.com/kotalco/community-api/api/handlers/filecoin"
	"github.com/kotalco/community-api/api/handlers/ipfs/ipfs_cluster_peer"
	"github.com/kotalco/community-api/api/handlers/ipfs/ipfs_peer"
	"github.com/kotalco/community-api/api/handlers/near"
	"github.com/kotalco/community-api/api/handlers/polkadot"
	"github.com/kotalco/community-api/api/handlers/shared"
	"github.com/kotalco/community-api/api/handlers/stacks"
)

// MapUrl abstracted function to map and register all the url for the application
func MapUrl(app *fiber.App) {
	api := app.Group("api")
	v1 := api.Group("v1")

	//subscription
	v1.Post("subscriptions/acknowledgment", subscription.Acknowledgement)
	v1.Use(middleware.IsSubscription)

	//users group
	v1.Post("sessions", user.SignIn)
	users := v1.Group("users")
	users.Post("/", user.SignUp)
	users.Post("/resend_email_verification", user.SendEmailVerification)
	users.Post("/forget_password", user.ForgetPassword)
	users.Post("/reset_password", user.ResetPassword)
	users.Post("/verify_email", user.VerifyEmail)

	users.Post("/change_password", middleware.JWTProtected, middleware.TFAProtected, user.ChangePassword)
	users.Post("/change_email", middleware.JWTProtected, middleware.TFAProtected, user.ChangeEmail)
	users.Get("/whoami", middleware.JWTProtected, middleware.TFAProtected, user.Whoami)

	users.Post("/totp", middleware.JWTProtected, user.CreateTOTP)
	users.Post("/totp/enable", middleware.JWTProtected, user.EnableTwoFactorAuth)
	users.Post("/totp/verify", middleware.JWTProtected, user.VerifyTOTP)
	users.Post("/totp/disable", middleware.JWTProtected, middleware.TFAProtected, user.DisableTwoFactorAuth)

	//workspace group
	workspaces := v1.Group("workspaces")
	workspaces.Use(middleware.JWTProtected, middleware.TFAProtected)
	workspaces.Post("/", workspace.Create)
	workspaces.Patch("/:id", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, middleware.IsWriter, workspace.Update)
	workspaces.Delete("/:id", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, middleware.IsWriter, workspace.Delete)
	workspaces.Get("/", workspace.GetByUserId)
	workspaces.Get("/:id", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, middleware.IsReader, workspace.GetById)
	workspaces.Post("/:id/members", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, workspace.AddMember)
	workspaces.Post("/:id/leave", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, workspace.Leave)
	workspaces.Delete("/:id/members/:user_id", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, workspace.RemoveMember)
	workspaces.Get("/:id/members", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, workspace.Members)
	workspaces.Patch("/:id/members/:user_id", workspace.ValidateWorkspaceExist, middleware.ValidateWorkspaceMembership, workspace.UpdateWorkspaceUser)

	//license group
	licenses := v1.Group("subscriptions")
	licenses.Get("/current", middleware.JWTProtected, middleware.TFAProtected, subscription.Current)

	//svc group
	svcGroup := v1.Group("/core/services")
	svcGroup.Get("/", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, svc.List)
	//svc group
	stsGroup := v1.Group("/core/statefulset")
	stsGroup.Get("/count", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, sts.Count)

	//endpoints group
	endpoints := v1.Group("endpoints")
	endpoints.Post("/", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, middleware.IsWriter, endpoint.Create)
	endpoints.Head("/", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, middleware.IsReader, endpoint.Count)
	endpoints.Get("/", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, middleware.IsReader, endpoint.List)
	endpoints.Get("/:name", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, middleware.IsReader, endpoint.Get)
	endpoints.Delete("/:name", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership, middleware.IsAdmin, endpoint.Delete)

	//settings group
	settingGroup := v1.Group("settings")
	settingGroup.Get("/", middleware.JWTProtected, middleware.TFAProtected, setting.Settings)
	settingGroup.Post("/domain", middleware.JWTProtected, middleware.TFAProtected, setting.ConfigureDomain)
	settingGroup.Post("/registration", middleware.JWTProtected, middleware.TFAProtected, setting.ConfigureRegistration)
	settingGroup.Get("/ip-address", middleware.JWTProtected, middleware.TFAProtected, setting.IPAddress)
	mapDeploymentUrl(v1)
}

func mapDeploymentUrl(v1 fiber.Router) {
	// chainlink group
	chainlinkGroup := v1.Group("chainlink", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	chainlinkNodes := chainlinkGroup.Group("nodes")

	chainlinkNodes.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, chainlink.Create)
	chainlinkNodes.Head("/", middleware.IsReader, chainlink.Count)
	chainlinkNodes.Get("/", middleware.IsReader, chainlink.List)
	chainlinkNodes.Get("/:name", middleware.IsReader, chainlink.ValidateNodeExist, chainlink.Get)
	chainlinkNodes.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	chainlinkNodes.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	chainlinkNodes.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	chainlinkNodes.Put("/:name", middleware.IsWriter, chainlink.ValidateNodeExist, chainlink.Update)
	chainlinkNodes.Delete("/:name", middleware.IsAdmin, chainlink.ValidateNodeExist, chainlink.Delete)

	//ethereum group
	ethereumGroup := v1.Group("ethereum", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	ethereumNodes := ethereumGroup.Group("nodes")
	ethereumNodes.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, ethereum.Create)
	ethereumNodes.Head("/", middleware.IsReader, ethereum.Count)
	ethereumNodes.Get("/", middleware.IsReader, ethereum.List)
	ethereumNodes.Get("/:name", middleware.IsReader, ethereum.ValidateNodeExist, ethereum.Get)
	ethereumNodes.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	ethereumNodes.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	ethereumNodes.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	ethereumNodes.Get("/:name/stats", middleware.IsReader, websocket.New(ethereum.Stats))
	ethereumNodes.Put("/:name", middleware.IsWriter, ethereum.ValidateNodeExist, ethereum.Update)
	ethereumNodes.Delete("/:name", middleware.IsAdmin, ethereum.ValidateNodeExist, ethereum.Delete)

	//core group
	coreGroup := v1.Group("core", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	//secret group
	secrets := coreGroup.Group("secrets")
	secrets.Post("/", middleware.IsWriter, secret.Create)
	secrets.Head("/", middleware.IsReader, secret.Count)
	secrets.Get("/", middleware.IsReader, secret.List)
	secrets.Get("/:name", middleware.IsReader, secret.ValidateSecretExist, secret.Get)
	secrets.Put("/:name", middleware.IsWriter, secret.ValidateSecretExist, secret.Update)
	secrets.Delete("/:name", middleware.IsAdmin, secret.ValidateSecretExist, secret.Delete)
	//storage class group
	storageClasses := coreGroup.Group("storageclasses", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	storageClasses.Post("/", middleware.IsWriter, storage_class.Create)
	storageClasses.Get("/", middleware.IsReader, storage_class.List)
	storageClasses.Get("/:name", middleware.IsReader, storage_class.ValidateStorageClassExist, storage_class.Get)
	storageClasses.Put("/:name", middleware.IsWriter, storage_class.ValidateStorageClassExist, storage_class.Update)
	storageClasses.Delete("/:name", middleware.IsAdmin, storage_class.ValidateStorageClassExist, storage_class.Delete)

	//ethereum2 group
	ethereum2 := v1.Group("ethereum2", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	//beaconnodes group
	beaconnodesGroup := ethereum2.Group("beaconnodes")
	beaconnodesGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, beacon_node.Create)
	beaconnodesGroup.Head("/", middleware.IsReader, beacon_node.Count)
	beaconnodesGroup.Get("/", middleware.IsReader, beacon_node.List)
	beaconnodesGroup.Get("/:name", middleware.IsReader, beacon_node.ValidateBeaconNodeExist, beacon_node.Get)
	beaconnodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	beaconnodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	beaconnodesGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	beaconnodesGroup.Get("/:name/stats", middleware.IsReader, websocket.New(beacon_node.Stats))
	beaconnodesGroup.Put("/:name", middleware.IsWriter, beacon_node.ValidateBeaconNodeExist, beacon_node.Update)
	beaconnodesGroup.Delete("/:name", middleware.IsAdmin, beacon_node.ValidateBeaconNodeExist, beacon_node.Delete)
	//validators group
	validatorsGroup := ethereum2.Group("validators", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	validatorsGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, validator.Create)
	validatorsGroup.Head("/", middleware.IsReader, validator.Count)
	validatorsGroup.Get("/", middleware.IsReader, validator.List)
	validatorsGroup.Get("/:name", middleware.IsReader, validator.ValidateValidatorExist, validator.Get)
	validatorsGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	validatorsGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	validatorsGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	validatorsGroup.Put("/:name", middleware.IsWriter, validator.ValidateValidatorExist, validator.Update)
	validatorsGroup.Delete("/:name", middleware.IsAdmin, validator.ValidateValidatorExist, validator.Delete)

	//filecoin group
	filecoinGroup := v1.Group("filecoin", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	filecoinNodes := filecoinGroup.Group("nodes")
	filecoinNodes.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, filecoin.Create)
	filecoinNodes.Head("/", middleware.IsReader, filecoin.Count)
	filecoinNodes.Get("/", middleware.IsReader, filecoin.List)
	filecoinNodes.Get("/:name", middleware.IsReader, filecoin.ValidateNodeExist, filecoin.Get)
	filecoinNodes.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	filecoinNodes.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	filecoinNodes.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	filecoinNodes.Put("/:name", middleware.IsWriter, filecoin.ValidateNodeExist, filecoin.Update)
	filecoinNodes.Delete("/:name", middleware.IsAdmin, filecoin.ValidateNodeExist, filecoin.Delete)

	//ipfs group
	ipfsGroup := v1.Group("ipfs", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	//ipfs peer group
	ipfsPeersGroup := ipfsGroup.Group("peers")
	ipfsPeersGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, ipfs_peer.Create)
	ipfsPeersGroup.Head("/", middleware.IsReader, ipfs_peer.Count)
	ipfsPeersGroup.Get("/", middleware.IsReader, ipfs_peer.List)
	ipfsPeersGroup.Get("/:name", middleware.IsReader, ipfs_peer.ValidatePeerExist, ipfs_peer.Get)
	ipfsPeersGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	ipfsPeersGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	ipfsPeersGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	ipfsPeersGroup.Get("/:name/stats", middleware.IsReader, websocket.New(ipfs_peer.Stats))
	ipfsPeersGroup.Put("/:name", middleware.IsWriter, ipfs_peer.ValidatePeerExist, ipfs_peer.Update)
	ipfsPeersGroup.Delete("/:name", middleware.IsAdmin, ipfs_peer.ValidatePeerExist, ipfs_peer.Delete)
	//ipfs peer group
	clusterpeersGroup := ipfsGroup.Group("clusterpeers", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	clusterpeersGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, ipfs_cluster_peer.Create)
	clusterpeersGroup.Head("/", middleware.IsReader, ipfs_cluster_peer.Count)
	clusterpeersGroup.Get("/", middleware.IsReader, ipfs_cluster_peer.List)
	clusterpeersGroup.Get("/:name", middleware.IsReader, ipfs_cluster_peer.ValidateClusterPeerExist, ipfs_cluster_peer.Get)
	clusterpeersGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	clusterpeersGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	clusterpeersGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	clusterpeersGroup.Put("/:name", middleware.IsWriter, ipfs_cluster_peer.ValidateClusterPeerExist, ipfs_cluster_peer.Update)
	clusterpeersGroup.Delete("/:name", middleware.IsAdmin, ipfs_cluster_peer.ValidateClusterPeerExist, ipfs_cluster_peer.Delete)

	//near group
	nearGroup := v1.Group("near", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	nearNodesGroup := nearGroup.Group("nodes")
	nearNodesGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, near.Create)
	nearNodesGroup.Head("/", middleware.IsReader, near.Count)
	nearNodesGroup.Get("/", middleware.IsReader, near.List)
	nearNodesGroup.Get("/:name", middleware.IsReader, near.ValidateNodeExist, near.Get)
	nearNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	nearNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	nearNodesGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	nearNodesGroup.Get("/:name/stats", middleware.IsReader, websocket.New(near.Stats))
	nearNodesGroup.Put("/:name", middleware.IsWriter, near.ValidateNodeExist, near.Update)
	nearNodesGroup.Delete("/:name", middleware.IsAdmin, near.ValidateNodeExist, near.Delete)

	polkadotGroup := v1.Group("polkadot", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	polkadotNodesGroup := polkadotGroup.Group("nodes")
	polkadotNodesGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, polkadot.Create)
	polkadotNodesGroup.Head("/", middleware.IsReader, polkadot.Count)
	polkadotNodesGroup.Get("/", middleware.IsReader, polkadot.List)
	polkadotNodesGroup.Get("/:name", middleware.IsReader, polkadot.ValidateNodeExist, polkadot.Get)
	polkadotNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	polkadotNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	polkadotNodesGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	polkadotNodesGroup.Get("/:name/stats", middleware.IsReader, websocket.New(polkadot.Stats))
	polkadotNodesGroup.Put("/:name", middleware.IsWriter, polkadot.ValidateNodeExist, polkadot.Update)
	polkadotNodesGroup.Delete("/:name", middleware.IsAdmin, polkadot.ValidateNodeExist, polkadot.Delete)

	bitcoinGroup := v1.Group("bitcoin", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	bitcoinNodesGroup := bitcoinGroup.Group("nodes")
	bitcoinNodesGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, bitcoin.Create)
	bitcoinNodesGroup.Head("/", middleware.IsReader, bitcoin.Count)
	bitcoinNodesGroup.Get("/", middleware.IsReader, bitcoin.List)
	bitcoinNodesGroup.Get("/:name", middleware.IsReader, bitcoin.ValidateNodeExist, bitcoin.Get)
	bitcoinNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	bitcoinNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	bitcoinNodesGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	bitcoinNodesGroup.Get("/:name/stats", middleware.IsReader, websocket.New(bitcoin.Stats))
	bitcoinNodesGroup.Put("/:name", middleware.IsWriter, bitcoin.ValidateNodeExist, bitcoin.Update)
	bitcoinNodesGroup.Delete("/:name", middleware.IsAdmin, bitcoin.ValidateNodeExist, bitcoin.Delete)

	stacksGroup := v1.Group("stacks", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	stacksNodesGroup := stacksGroup.Group("nodes")
	stacksNodesGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, stacks.Create)
	stacksNodesGroup.Head("/", middleware.IsReader, stacks.Count)
	stacksNodesGroup.Get("/", middleware.IsReader, stacks.List)
	stacksNodesGroup.Get("/:name", middleware.IsReader, stacks.ValidateNodeExist, stacks.Get)
	stacksNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	stacksNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	stacksNodesGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	stacksNodesGroup.Put("/:name", middleware.IsWriter, stacks.ValidateNodeExist, stacks.Update)
	stacksNodesGroup.Delete("/:name", middleware.IsAdmin, stacks.ValidateNodeExist, stacks.Delete)

	aptosGroup := v1.Group("aptos", middleware.JWTProtected, middleware.TFAProtected, middleware.WorkspaceProtected, middleware.ValidateWorkspaceMembership)
	aptosNodesGroup := aptosGroup.Group("nodes")
	aptosNodesGroup.Post("/", middleware.IsWriter, middleware.NodesLimitProtected, aptos.Create)
	aptosNodesGroup.Head("/", middleware.IsReader, aptos.Count)
	aptosNodesGroup.Get("/", middleware.IsReader, aptos.List)
	aptosNodesGroup.Get("/:name", middleware.IsReader, aptos.ValidateNodeExist, aptos.Get)
	aptosNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	aptosNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	aptosNodesGroup.Get("/:name/metrics", middleware.IsReader, websocket.New(shared.Metrics))
	aptosNodesGroup.Put("/:name", middleware.IsWriter, aptos.ValidateNodeExist, aptos.Update)
	aptosNodesGroup.Delete("/:name", middleware.IsAdmin, aptos.ValidateNodeExist, aptos.Delete)

}
