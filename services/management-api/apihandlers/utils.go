package apihandlers

func (h *HttpEndpoints) isInstanceAllowed(instanceID string) bool {
	for _, id := range h.allowedInstanceIDs {
		if id == instanceID {
			return true
		}
	}
	return false
}
