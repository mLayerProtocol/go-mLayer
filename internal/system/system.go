package system

import (
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
)

func RegisterNodeInterest(validator entities.PublicKeyString, ids []string, modelType entities.EntityModel, ownInterest *[]string) ( err error) {
	if ownInterest != nil {
		topicsOfInterest := []string{}
		for _, id := range ids {
			if int, _ := dsquery.IsInterestedIn(id, modelType); !int {
				// if err != nil && !dsquery.IsErrorNotFound(err) {

				// }
				continue
			}
			if intersted, _ := dsquery.IsNodeInterested(string(validator), id, modelType); !intersted {
				// notify node that i am interested
				topicsOfInterest = append(topicsOfInterest, id)
			}
		}
		*ownInterest = topicsOfInterest
	}
	 err = dsquery.SetNodeInterest(string(validator), ids, modelType)
	
	return err

}


