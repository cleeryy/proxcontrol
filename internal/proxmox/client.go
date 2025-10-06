package proxmox

import (
    "crypto/tls"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"
    "encoding/json"
)

type Client struct {
    BaseURL    string
    TokenID    string
    Secret     string
    Node       string
    HTTPClient *http.Client
    AllowedVMs map[int]bool
}

type VMStatus struct {
    Status string  `json:"status"`
    Name   string  `json:"name"`
    VMID   int     `json:"vmid"`
    Uptime int64   `json:"uptime"`
    CPU    float64 `json:"cpu"`
    Mem    int64   `json:"mem"`
    MaxMem int64   `json:"maxmem"`
}

type APIResponse struct {
    Data json.RawMessage `json:"data"`
}

// NewClient crée un nouveau client Proxmox
func NewClient(baseURL, tokenID, secret, node string, allowedVMs []int) *Client {
    vmMap := make(map[int]bool)
    for _, vmid := range allowedVMs {
        vmMap[vmid] = true
    }

    return &Client{
        BaseURL: strings.TrimSuffix(baseURL, "/"),
        TokenID: tokenID,
        Secret:  secret,
        Node:    node,
        AllowedVMs: vmMap,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true,
                },
            },
        },
    }
}

// IsVMAllowed vérifie si une VM est dans la whitelist
func (c *Client) IsVMAllowed(vmid int) bool {
    return c.AllowedVMs[vmid]
}

// doRequest effectue une requête HTTP à l'API Proxmox
func (c *Client) doRequest(method, endpoint string, params map[string]string) ([]byte, error) {
    reqURL := c.BaseURL + endpoint

    var reqBody io.Reader
    
    // Pour POST/PUT avec paramètres, utiliser x-www-form-urlencoded
    if (method == "POST" || method == "PUT") && params != nil && len(params) > 0 {
        data := url.Values{}
        for key, value := range params {
            data.Set(key, value)
        }
        reqBody = strings.NewReader(data.Encode())
    }

    req, err := http.NewRequest(method, reqURL, reqBody)
    if err != nil {
        return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
    }

    // Header d'authentification avec le token API
    authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
    req.Header.Set("Authorization", authHeader)
    
    // Content-Type pour POST/PUT
    if (method == "POST" || method == "PUT") && params != nil && len(params) > 0 {
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("erreur lors de la requête HTTP: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %w", err)
    }

    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("erreur API Proxmox: %d - %s", resp.StatusCode, string(respBody))
    }

    return respBody, nil
}

// GetVMStatus récupère l'état actuel d'une VM
func (c *Client) GetVMStatus(vmid int) (*VMStatus, error) {
    if !c.IsVMAllowed(vmid) {
        return nil, fmt.Errorf("VM %d non autorisée", vmid)
    }

    endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", c.Node, vmid)
    data, err := c.doRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }

    var response APIResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("erreur de décodage JSON: %w", err)
    }

    var status VMStatus
    if err := json.Unmarshal(response.Data, &status); err != nil {
        return nil, fmt.Errorf("erreur de décodage des données VM: %w", err)
    }

    return &status, nil
}

// StartVM démarre une VM
func (c *Client) StartVM(vmid int) error {
    if !c.IsVMAllowed(vmid) {
        return fmt.Errorf("VM %d non autorisée", vmid)
    }

    endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", c.Node, vmid)
    _, err := c.doRequest("POST", endpoint, nil)
    return err
}

// StopVM arrête une VM (arrêt forcé)
func (c *Client) StopVM(vmid int) error {
    if !c.IsVMAllowed(vmid) {
        return fmt.Errorf("VM %d non autorisée", vmid)
    }

    endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", c.Node, vmid)
    _, err := c.doRequest("POST", endpoint, nil)
    return err
}

// ShutdownVM éteint proprement une VM (graceful shutdown)
func (c *Client) ShutdownVM(vmid int) error {
    if !c.IsVMAllowed(vmid) {
        return fmt.Errorf("VM %d non autorisée", vmid)
    }

    endpoint := fmt.Sprintf("/nodes/%s/qemu/%d/status/shutdown", c.Node, vmid)
    _, err := c.doRequest("POST", endpoint, nil)
    return err
}

// ParseAllowedVMs convertit une string CSV en slice d'entiers
func ParseAllowedVMs(vmsStr string) ([]int, error) {
    if vmsStr == "" {
        return []int{}, nil
    }

    parts := strings.Split(vmsStr, ",")
    vms := make([]int, 0, len(parts))

    for _, part := range parts {
        vmid, err := strconv.Atoi(strings.TrimSpace(part))
        if err != nil {
            return nil, fmt.Errorf("ID VM invalide: %s", part)
        }
        vms = append(vms, vmid)
    }

    return vms, nil
}

// VM représente une VM Proxmox
type VM struct {
    VMID   int    `json:"vmid"`
    Name   string `json:"name"`
    Status string `json:"status"`
    Node   string `json:"node"`
}

// ListVMs récupère la liste de toutes les VMs du cluster
func (c *Client) ListVMs() ([]VM, error) {
    endpoint := "/cluster/resources?type=vm"
    data, err := c.doRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }

    var response APIResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("erreur de décodage JSON: %w", err)
    }

    var vms []VM
    if err := json.Unmarshal(response.Data, &vms); err != nil {
        return nil, fmt.Errorf("erreur de décodage des VMs: %w", err)
    }

    // Filtrer uniquement les VMs autorisées
    allowedVMs := make([]VM, 0)
    for _, vm := range vms {
        if c.IsVMAllowed(vm.VMID) {
            allowedVMs = append(allowedVMs, vm)
        }
    }

    return allowedVMs, nil
}

// FindVMByName cherche une VM par son nom
func (c *Client) FindVMByName(name string) (*VM, error) {
    vms, err := c.ListVMs()
    if err != nil {
        return nil, err
    }

    nameLower := strings.ToLower(name)
    for _, vm := range vms {
        if strings.ToLower(vm.Name) == nameLower {
            return &vm, nil
        }
    }

    return nil, fmt.Errorf("VM '%s' non trouvée", name)
}

