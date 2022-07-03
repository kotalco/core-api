package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/kotalco/api/api/handlers/chainlink"
	"github.com/kotalco/api/api/handlers/core/secret"
	"github.com/kotalco/api/api/handlers/core/storage_class"
	"github.com/kotalco/api/api/handlers/ethereum"
	"github.com/kotalco/api/api/handlers/ethereum2/beacon_node"
	"github.com/kotalco/api/api/handlers/ethereum2/validator"
	"github.com/kotalco/api/api/handlers/filecoin"
	"github.com/kotalco/api/api/handlers/ipfs/ipfs_cluster_peer"
	"github.com/kotalco/api/api/handlers/ipfs/ipfs_peer"
	"github.com/kotalco/api/api/handlers/near"
	"github.com/kotalco/api/api/handlers/polkadot"
	"github.com/kotalco/api/api/handlers/shared"
	"github.com/kotalco/cloud-api/api/handler/user"
	"github.com/kotalco/cloud-api/api/handler/workspace"
	"github.com/kotalco/cloud-api/pkg/middleware"
)

// MapUrl abstracted function to map and register all the url for the application
func MapUrl(app *fiber.App) {
	api := app.Group("api")
	v1 := api.Group("v1")
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
	workspaces.Patch("/:id", middleware.IsWorkspace, middleware.IsWriter, workspace.Update)
	workspaces.Delete("/:id", middleware.IsWorkspace, middleware.IsWriter, workspace.Delete)
	workspaces.Get("/", workspace.GetByUserId)
	workspaces.Get("/:id", middleware.IsWorkspace, middleware.IsReader, workspace.GetById)
	workspaces.Post("/:id/members", middleware.IsWorkspace, workspace.AddMember)
	workspaces.Post("/:id/leave", middleware.IsWorkspace, workspace.Leave)
	workspaces.Delete("/:id/members/:user_id", middleware.IsWorkspace, workspace.RemoveMember)
	workspaces.Get("/:id/members", middleware.IsWorkspace, workspace.Members)
	workspaces.Patch("/:id/members/:user_id", middleware.IsWorkspace, workspace.UpdateWorkspaceUser)

	//community routes
	mapDeploymentUrl(v1)
}

func mapDeploymentUrl(v1 fiber.Router) {
	v1.Use(middleware.JWTProtected, middleware.TFAProtected)

	// chainlink group
	chainlinkGroup := v1.Group("chainlink")
	chainlinkNodes := chainlinkGroup.Group("nodes")

	chainlinkNodes.Post("/", middleware.IsAdmin, chainlink.Create)
	chainlinkNodes.Head("/", middleware.IsReader, chainlink.Count)
	chainlinkNodes.Get("/", middleware.IsReader, chainlink.List)
	chainlinkNodes.Get("/:name", middleware.IsReader, chainlink.ValidateNodeExist, chainlink.Get)
	chainlinkNodes.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	chainlinkNodes.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	chainlinkNodes.Put("/:name", middleware.IsWriter, chainlink.ValidateNodeExist, chainlink.Update)
	chainlinkNodes.Delete("/:name", middleware.IsAdmin, chainlink.ValidateNodeExist, chainlink.Delete)

	//ethereum group
	ethereumGroup := v1.Group("ethereum")
	ethereumNodes := ethereumGroup.Group("nodes")
	ethereumNodes.Post("/", middleware.IsWriter, ethereum.Create)
	ethereumNodes.Head("/", middleware.IsReader, ethereum.Count)
	ethereumNodes.Get("/", middleware.IsReader, ethereum.List)
	ethereumNodes.Get("/:name", middleware.IsReader, ethereum.ValidateNodeExist, ethereum.Get)
	ethereumNodes.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	ethereumNodes.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	ethereumNodes.Get("/:name/stats", middleware.IsReader, websocket.New(ethereum.Stats))
	ethereumNodes.Put("/:name", middleware.IsWriter, ethereum.ValidateNodeExist, ethereum.Update)
	ethereumNodes.Delete("/:name", middleware.IsAdmin, ethereum.ValidateNodeExist, ethereum.Delete)

	//core group
	coreGroup := v1.Group("core")
	//secret group
	secrets := coreGroup.Group("secrets")
	secrets.Post("/", middleware.IsWriter, secret.Create)
	secrets.Head("/", middleware.IsReader, secret.Count)
	secrets.Get("/", middleware.IsReader, secret.List)
	secrets.Get("/:name", middleware.IsReader, secret.ValidateSecretExist, secret.Get)
	secrets.Put("/:name", middleware.IsWriter, secret.ValidateSecretExist, secret.Update)
	secrets.Delete("/:name", middleware.IsAdmin, secret.ValidateSecretExist, secret.Delete)
	//storage class group
	storageClasses := coreGroup.Group("storageclasses")
	storageClasses.Post("/", middleware.IsWriter, storage_class.Create)
	storageClasses.Get("/", middleware.IsReader, storage_class.List)
	storageClasses.Get("/:name", middleware.IsReader, storage_class.ValidateStorageClassExist, storage_class.Get)
	storageClasses.Put("/:name", middleware.IsWriter, storage_class.ValidateStorageClassExist, storage_class.Update)
	storageClasses.Delete("/:name", middleware.IsAdmin, storage_class.ValidateStorageClassExist, storage_class.Delete)

	//ethereum2 group
	ethereum2 := v1.Group("ethereum2")
	//beaconnodes group
	beaconnodesGroup := ethereum2.Group("beaconnodes")
	beaconnodesGroup.Post("/", middleware.IsWriter, beacon_node.Create)
	beaconnodesGroup.Head("/", middleware.IsReader, beacon_node.Count)
	beaconnodesGroup.Get("/", middleware.IsReader, beacon_node.List)
	beaconnodesGroup.Get("/:name", middleware.IsReader, beacon_node.ValidateBeaconNodeExist, beacon_node.Get)
	beaconnodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	beaconnodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	beaconnodesGroup.Put("/:name", middleware.IsWriter, beacon_node.ValidateBeaconNodeExist, beacon_node.Update)
	beaconnodesGroup.Delete("/:name", middleware.IsAdmin, beacon_node.ValidateBeaconNodeExist, beacon_node.Delete)
	//validators group
	validatorsGroup := ethereum2.Group("validators")
	validatorsGroup.Post("/", middleware.IsWriter, validator.Create)
	validatorsGroup.Head("/", middleware.IsReader, validator.Count)
	validatorsGroup.Get("/", middleware.IsReader, validator.List)
	validatorsGroup.Get("/:name", middleware.IsReader, validator.ValidateValidatorExist, validator.Get)
	validatorsGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	validatorsGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	validatorsGroup.Put("/:name", middleware.IsWriter, validator.ValidateValidatorExist, validator.Update)
	validatorsGroup.Delete("/:name", middleware.IsAdmin, validator.ValidateValidatorExist, validator.Delete)

	//filecoin group
	filecoinGroup := v1.Group("filecoin")
	filecoinNodes := filecoinGroup.Group("nodes")
	filecoinNodes.Post("/", middleware.IsWriter, filecoin.Create)
	filecoinNodes.Head("/", middleware.IsReader, filecoin.Count)
	filecoinNodes.Get("/", middleware.IsReader, filecoin.List)
	filecoinNodes.Get("/:name", middleware.IsReader, filecoin.ValidateNodeExist, filecoin.Get)
	filecoinNodes.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	filecoinNodes.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	filecoinNodes.Put("/:name", middleware.IsWriter, filecoin.ValidateNodeExist, filecoin.Update)
	filecoinNodes.Delete("/:name", middleware.IsAdmin, filecoin.ValidateNodeExist, filecoin.Delete)

	//ipfs group
	ipfsGroup := v1.Group("ipfs")
	//ipfs peer group
	ipfsPeersGroup := ipfsGroup.Group("peers")
	ipfsPeersGroup.Post("/", middleware.IsWriter, ipfs_peer.Create)
	ipfsPeersGroup.Head("/", middleware.IsReader, ipfs_peer.Count)
	ipfsPeersGroup.Get("/", middleware.IsReader, ipfs_peer.List)
	ipfsPeersGroup.Get("/:name", middleware.IsReader, ipfs_peer.ValidatePeerExist, ipfs_peer.Get)
	ipfsPeersGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	ipfsPeersGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	ipfsPeersGroup.Put("/:name", middleware.IsWriter, ipfs_peer.ValidatePeerExist, ipfs_peer.Update)
	ipfsPeersGroup.Delete("/:name", middleware.IsAdmin, ipfs_peer.ValidatePeerExist, ipfs_peer.Delete)
	//ipfs peer group
	clusterpeersGroup := ipfsGroup.Group("clusterpeers")
	clusterpeersGroup.Post("/", middleware.IsWriter, ipfs_cluster_peer.Create)
	clusterpeersGroup.Head("/", middleware.IsReader, ipfs_cluster_peer.Count)
	clusterpeersGroup.Get("/", middleware.IsReader, ipfs_cluster_peer.List)
	clusterpeersGroup.Get("/:name", middleware.IsReader, ipfs_cluster_peer.ValidateClusterPeerExist, ipfs_cluster_peer.Get)
	clusterpeersGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	clusterpeersGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	clusterpeersGroup.Put("/:name", middleware.IsWriter, ipfs_cluster_peer.ValidateClusterPeerExist, ipfs_cluster_peer.Update)
	clusterpeersGroup.Delete("/:name", middleware.IsAdmin, ipfs_cluster_peer.ValidateClusterPeerExist, ipfs_cluster_peer.Delete)

	//near group
	nearGroup := v1.Group("near")
	nearNodesGroup := nearGroup.Group("nodes")
	nearNodesGroup.Post("/", middleware.IsWriter, near.Create)
	nearNodesGroup.Head("/", middleware.IsReader, near.Count)
	nearNodesGroup.Get("/", middleware.IsReader, near.List)
	nearNodesGroup.Get("/:name", middleware.IsReader, near.ValidateNodeExist, near.Get)
	nearNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	nearNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	nearNodesGroup.Get("/:name/stats", middleware.IsReader, websocket.New(near.Stats))
	nearNodesGroup.Put("/:name", middleware.IsWriter, near.ValidateNodeExist, near.Update)
	nearNodesGroup.Delete("/:name", middleware.IsAdmin, near.ValidateNodeExist, near.Delete)

	polkadotGroup := v1.Group("polkadot")
	polkadotNodesGroup := polkadotGroup.Group("nodes")
	polkadotNodesGroup.Post("/", middleware.IsWriter, polkadot.Create)
	polkadotNodesGroup.Head("/", middleware.IsReader, polkadot.Count)
	polkadotNodesGroup.Get("/", middleware.IsReader, polkadot.List)
	polkadotNodesGroup.Get("/:name", middleware.IsReader, polkadot.ValidateNodeExist, polkadot.Get)
	polkadotNodesGroup.Get("/:name/logs", middleware.IsReader, websocket.New(shared.Logger))
	polkadotNodesGroup.Get("/:name/status", middleware.IsReader, websocket.New(shared.Status))
	polkadotNodesGroup.Get("/:name/stats", middleware.IsReader, websocket.New(polkadot.Stats))
	polkadotNodesGroup.Put("/:name", middleware.IsWriter, polkadot.ValidateNodeExist, polkadot.Update)
	polkadotNodesGroup.Delete("/:name", middleware.IsAdmin, polkadot.ValidateNodeExist, polkadot.Delete)

}
