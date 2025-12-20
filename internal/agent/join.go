package agent

import (
	"encoding/json"
	"net/http"
)

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ Token string }
	json.NewDecoder(r.Body).Decode(&req)

	// nodeInfo := getNodeInfo()

	// resp, err := manager.RegisterNode(req.Token, nodeInfo)
	// if err != nil {
	// 	http.Error(w, err.Error(), 403)
	// 	return
	// }

	// cert.WriteNodeCert(resp.Cert)

	// lxd.Install()
	// lxd.JoinCluster(resp.ClusterEndpoint, resp.Cert)

	w.WriteHeader(200)
}
