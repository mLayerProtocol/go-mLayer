package client

// func SyncAgent(req *entities.SyncRequest, clientPayload *entities.ClientPayload) (entities.SyncResponse, error) {
// func SyncAgent(req *entities.SyncRequest, clientPayload *entities.ClientPayload) (entities.SyncResponse, error) {
// 	//agentAuthState, _ := ValidateClientPayload(clientPayload)

// 	// if agentAuthState == nil || agentAuthState.Priviledge == 0 {
// 	// 	return nil, apperror.Unauthorized("Agent not authorized")
// 	// }
// 	query.GetSubscriptionsByBlock(entities.Subscription{}, 1, 3, true)
// 	syncResponse := entities.SyncResponse{}
// 	useTime := true
// 	if req.Interval.FromBlock > 0 {
// 		useTime = false
// 	}
// 	if useTime {
// 		syncResponse.TimeFrame.ToTime = req.Interval.ToTime
// 		syncResponse.TimeFrame.FromTime = req.Interval.FromTime
// 	} else {
// 		syncResponse.TimeFrame.ToBlock = req.Interval.ToBlock
// 		syncResponse.TimeFrame.FromBlock = req.Interval.FromBlock
// 	}

// 	query.GetSubscriptionsByBlock(entities.Subscription{}, 1, 3, true)
// 	//syncResponse.Topics =
// 	return syncResponse, nil

// 	// getJoins := query.GetSubscriptionsByBlock(entities.Subscription{}, 1, 3, true)

// 	// syncResponse.Topics.Joins
// 	// err := query.GetMany(models.AuthorizationState{
// 	// 	Authorization: *auth,
// 	// }, &authState)
// 	// if err != nil {
// 	// 	if err == gorm.ErrRecordNotFound {
// 	// 		return nil, nil
// 	// 	}
// 	// 	return nil, err
// 	// }
// 	// return &authState, nil
// }