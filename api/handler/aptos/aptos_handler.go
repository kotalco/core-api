package aptos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/kotalco/core-api/core/aptos"
	"github.com/kotalco/core-api/k8s"
	restErrors "github.com/kotalco/core-api/pkg/errors"
	"github.com/kotalco/core-api/pkg/pagination"
	"github.com/kotalco/core-api/pkg/responder"
	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type request struct {
	url  string
	name string
}
type result struct {
	err  error
	name string
	data []byte
}

const (
	nameKeyword = "name"
)

var (
	service   = aptos.NewAptosService()
	k8sClient = k8s.NewClientService()
)

// Get returns a single aptos node by name
func Get(c *fiber.Ctx) error {
	node := c.Locals("node").(aptosv1alpha1.Node)
	return c.JSON(responder.NewResponse(new(aptos.AptosDto).FromAptosNode(node)))
}

// List returns all aptos nodes
func List(c *fiber.Ctx) error {
	// default page to 0
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	nodeList, err := service.List(c.Locals("namespace").(string))
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	start, end := pagination.Page(uint(len(nodeList.Items)), uint(page), uint(limit))
	sort.Slice(nodeList.Items[:], func(i, j int) bool {
		return nodeList.Items[j].CreationTimestamp.Before(&nodeList.Items[i].CreationTimestamp)
	})

	c.Set("Access-Control-Expose-Headers", "X-Total-Count")
	c.Set("X-Total-Count", fmt.Sprintf("%d", len(nodeList.Items)))

	return c.Status(http.StatusOK).JSON(responder.NewResponse(new(aptos.AptosListDto).FromAptosNode(nodeList.Items[start:end])))
}

// Create created aptos node from given specs
func Create(c *fiber.Ctx) error {
	dto := new(aptos.AptosDto)
	if err := c.BodyParser(dto); err != nil {
		badReqErr := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReqErr.StatusCode()).JSON(badReqErr)
	}

	dto.Namespace = c.Locals("namespace").(string)
	if err := dto.MetaDataDto.Validate(); err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	node, err := service.Create(*dto)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}
	return c.Status(http.StatusCreated).JSON(responder.NewResponse(new(aptos.AptosDto).FromAptosNode(node)))
}

// Update updates a single aptos node by name from spec
func Update(c *fiber.Ctx) error {
	dto := new(aptos.AptosDto)
	if err := c.BodyParser(dto); err != nil {
		badReq := restErrors.NewBadRequestError("invalid request body")
		return c.Status(badReq.StatusCode()).JSON(badReq)
	}

	node := c.Locals("node").(aptosv1alpha1.Node)

	err := service.Update(*dto, &node)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	return c.Status(http.StatusOK).JSON(responder.NewResponse(new(aptos.AptosDto).FromAptosNode(node)))
}

// Count returns total number of nodes
func Count(c *fiber.Ctx) error {
	length, err := service.Count(c.Locals("namespace").(string))
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	c.Set("Access-Control-Expose-Headers", "X-Total-Count")
	c.Set("X-Total-Count", fmt.Sprintf("%d", length))

	return c.SendStatus(http.StatusOK)
}

// Delete a single aptos node by name
func Delete(c *fiber.Ctx) error {
	node := c.Locals("node").(aptosv1alpha1.Node)

	err := service.Delete(&node)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	return c.SendStatus(http.StatusNoContent)
}

// Stats returns a websocket that emits aptos stats
func Stats(c *websocket.Conn) {
	defer c.Close()

	name := c.Params("name")
	node := &aptosv1alpha1.Node{}
	nameSpacedName := types.NamespacedName{
		Namespace: c.Locals("namespace").(string),
		Name:      name,
	}
	err := k8sClient.Get(context.Background(), nameSpacedName, node)
	if err != nil {
		if errors.IsNotFound(err) {
			c.WriteJSON(fiber.Map{
				"error": fmt.Sprintf("node by name %s doesn't exist", name),
			})
			return
		}
		c.WriteJSON(fiber.Map{
			"error": err.Error(),
		})
		return
	}

	if !node.Spec.API {
		c.WriteJSON(fiber.Map{
			"error": "node api is not enabled",
		})
		return
	}
	reqCount := 2

	for {
		jobs := make(chan request, reqCount)
		results := make(chan result, reqCount)

		for i := 0; i < reqCount; i++ {
			go worker(jobs, results)
		}

		jobs <- request{name: "ledgerInfo", url: fmt.Sprintf("http://%s.%s:%d/v1", node.Name, nameSpacedName.Namespace, node.Spec.APIPort)}
		jobs <- request{name: "nodeInspection", url: fmt.Sprintf("http://%s.%s:%d/json_metrics", node.Name, nameSpacedName.Namespace, node.Spec.MetricsPort)}

		close(jobs)

		var aptosStatResponseDto struct {
			CurrentBlock string  `json:"currentBlock"`
			PeerCount    float64 `json:"peerCount"`
		}

		newAptosResponse := aptosStatResponseDto

		for i := 0; i < reqCount; i++ {
			resp := <-results
			if resp.err != nil {
				c.WriteJSON(fiber.Map{
					"error": err.Error(),
				})
				continue
			}

			switch resp.name {
			case "ledgerInfo":
				var responseBody map[string]interface{}
				err = json.Unmarshal(resp.data, &responseBody)
				if err != nil {
					c.WriteJSON(fiber.Map{
						"error": err.Error(),
					})
					break
				}
				newAptosResponse.CurrentBlock = responseBody["block_height"].(string)
				break
			case "nodeInspection":

				var responseBody map[string]interface{}
				err = json.Unmarshal(resp.data, &responseBody)
				if err != nil {
					c.WriteJSON(fiber.Map{
						"error": err.Error(),
					})
					break
				}
				for key, value := range responseBody {
					if strings.HasPrefix(key, "aptos_connections.outbound.Public") {
						newAptosResponse.PeerCount = value.(float64)
						break
					}
				}
				break
			}

		}
		close(results)

		err := c.WriteJSON(newAptosResponse)
		if err != nil {
			return
		}

		time.Sleep(time.Second * 3)
	}
}

// worker is a  collection of threads for the aptos node stats
func worker(jobs <-chan request, results chan<- result) {
	chanRes := result{}
	for job := range jobs {
		chanRes.name = job.name

		client := http.Client{
			Timeout: 20 * time.Second,
		}
		req, err := http.NewRequest(http.MethodGet, job.url, bytes.NewReader([]byte(nil)))
		if err != nil {
			chanRes.err = err
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			chanRes.err = err
			return
		}

		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			chanRes.err = err
			return
		}
		chanRes.data = responseData
		results <- chanRes
	}
}

func ValidateNodeExist(c *fiber.Ctx) error {
	nameSpacedName := types.NamespacedName{
		Name:      c.Params(nameKeyword),
		Namespace: c.Locals("namespace").(string),
	}

	node, err := service.Get(nameSpacedName)
	if err != nil {
		return c.Status(err.StatusCode()).JSON(err)
	}

	c.Locals("node", node)
	return c.Next()
}
