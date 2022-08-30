package filestore_test

import (
	"testing"

	"github.com/monadicstack/filestore"
	"github.com/stretchr/testify/suite"
)

type PathTestSuite struct {
	suite.Suite
}

func (s *PathTestSuite) TestChangeExtension() {
	// Remove extension
	s.Require().Equal("", filestore.ChangeExtension("", ""))
	s.Require().Equal("foo", filestore.ChangeExtension("foo", ""))
	s.Require().Equal("foo", filestore.ChangeExtension("foo.", ""))
	s.Require().Equal("foo.bar", filestore.ChangeExtension("foo.bar.", ""))

	// For all examples, make sure that we accept extensions with and without the "."

	s.Require().Equal(".txt", filestore.ChangeExtension("", "txt"))
	s.Require().Equal(".txt", filestore.ChangeExtension("", ".txt"))

	s.Require().Equal("foo.txt", filestore.ChangeExtension("foo.", "txt"))
	s.Require().Equal("foo.txt", filestore.ChangeExtension("foo.", ".txt"))

	s.Require().Equal("foo.txt", filestore.ChangeExtension("foo.txt", "txt"))
	s.Require().Equal("foo.txt", filestore.ChangeExtension("foo.txt", ".txt"))

	s.Require().Equal("foo.txt", filestore.ChangeExtension("foo.png", "txt"))
	s.Require().Equal("foo.txt", filestore.ChangeExtension("foo.png", ".txt"))

	s.Require().Equal("foo.bar.txt", filestore.ChangeExtension("foo.bar.png", "txt"))
	s.Require().Equal("foo.bar.txt", filestore.ChangeExtension("foo.bar.png", ".txt"))

	s.Require().Equal("txt.txt.jpeg", filestore.ChangeExtension("txt.txt.txt", "jpeg"))
	s.Require().Equal("txt.txt.jpeg", filestore.ChangeExtension("txt.txt.txt", ".jpeg"))

	s.Require().Equal("a.super-long", filestore.ChangeExtension("a.b", "super-long"))
	s.Require().Equal("a.super-long", filestore.ChangeExtension("a.b", ".super-long"))

	s.Require().Equal("a.multi.dot.ext", filestore.ChangeExtension("a.b", "multi.dot.ext"))
	s.Require().Equal("a.multi.dot.ext", filestore.ChangeExtension("a.b", ".multi.dot.ext"))

	s.Require().Equal("a.super-🍺", filestore.ChangeExtension("a.b", "super-🍺"))
	s.Require().Equal("a.super-🍺", filestore.ChangeExtension("a.b", ".super-🍺"))
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, &PathTestSuite{})
}
