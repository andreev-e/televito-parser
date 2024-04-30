package common

import Dbmethods "televito-parser/dbmethods"

func restoreDeletedAdds(class string, sourceList map[uint32]interface{}) {
	ids := make([]uint32, 0)

	for key := range sourceList {
		ids = append(ids, key)
	}

	Dbmethods.RestoreTrashedAdds(ids, class)
}
