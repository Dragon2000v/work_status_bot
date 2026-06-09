package flows

import "testing"

func TestMongoRepositoryImplementsRepository(t *testing.T) {
	var _ Repository = (*MongoRepository)(nil)
}
