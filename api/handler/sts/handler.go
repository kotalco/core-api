package sts

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kotalco/cloud-api/pkg/k8s/statefulset"
	"github.com/kotalco/community-api/pkg/shared"
	"net/http"
)

var statefulSetService = statefulset.NewService()

func Count(c *fiber.Ctx) error {
	list, err := statefulSetService.List(c.Locals("namespace").(string))
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	dto := statefulset.CountResponseDto{}

	for _, v := range list.Items {
		switch v.Labels["kotal.io/protocol"] {
		case statefulset.StatFulSetProtocolList.Bitcoin:
			dto.Bitcoin++
			break
		case statefulset.StatFulSetProtocolList.Chainlink:
			dto.Chainlink++
			break
		case statefulset.StatFulSetProtocolList.Ethereum:
			dto.Ethereum++
			break
		case statefulset.StatFulSetProtocolList.Ethereum2:
			dto.Ethereum2++
			break
		case statefulset.StatFulSetProtocolList.Filecoin:
			dto.Filecoin++
			break
		case statefulset.StatFulSetProtocolList.Ipfs:
			dto.Ipfs++
			break
		case statefulset.StatFulSetProtocolList.Near:
			dto.Near++
			break
		case statefulset.StatFulSetProtocolList.Polkadot:
			dto.Polkadot++
			break
		case statefulset.StatFulSetProtocolList.Stacks:
			dto.Stacks++
			break
		}
	}

	return c.Status(http.StatusOK).JSON(shared.NewResponse(dto))
}
