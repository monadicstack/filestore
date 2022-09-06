package filestore_test

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/monadicstack/filestore"
	"github.com/stretchr/testify/suite"
)

type DiskTestSuite struct {
	suite.Suite
	tempDirPath string
}

func TestDiskTestSuite(t *testing.T) {
	suite.Run(t, &DiskTestSuite{})
}

func (s *DiskTestSuite) SetupTest() {
	dir := "testdata/inner1/lebowski"
	_ = os.RemoveAll(dir) // start from a fresh slate even if teardown didn't happen last time.

	// The "lebowski/" directory goes inside of the "inner1/" directory because none
	// of our tests explicitly test that folder's contents, so we can make that dynamic
	// as we want without accidentally breaking other tests.
	//
	// Also, the lebowski/ directory and the .lebowski files are in .gitignore, so we won't
	// accidentally check these in no matter where they end up.
	err := os.Mkdir(dir, 0755)
	s.Require().NoError(err, "Should be able to make temp directory during test setup...")

	s.Require().NoError(os.WriteFile(path.Join(dir, "1.lebowski"), []byte("jeff"), 0666))
	s.Require().NoError(os.WriteFile(path.Join(dir, "2.lebowski"), []byte("walter"), 0666))
	s.Require().NoError(os.WriteFile(path.Join(dir, "3.lebowski"), []byte("donnie"), 0666))
	s.Require().NoError(os.WriteFile(path.Join(dir, "4.lebowski"), []byte("maude"), 0666))

	s.Require().NoError(os.Mkdir(path.Join(dir, "dude"), 0755))
	s.Require().NoError(os.Mkdir(path.Join(dir, "duderino"), 0755))
	s.Require().NoError(os.WriteFile(path.Join(dir, "duderino", "5.lebowski"), []byte("jackie"), 0666))
	s.Require().NoError(os.WriteFile(path.Join(dir, "duderino", "6.lebowski"), []byte("nihilist"), 0666))

	s.tempDirPath = dir
}

func (s *DiskTestSuite) TearDownTest() {
	_ = os.RemoveAll(s.tempDirPath)
}

func (s *DiskTestSuite) TestStat() {
	fs := filestore.Disk("testdata")

	info, err := fs.Stat("hello.txt")
	s.Require().NoError(err, "Running 'stat' on valid file should not give an error")
	s.Require().Equal("hello.txt", info.Name())
	s.Require().Equal(int64(12), info.Size())

	info, err = fs.Stat("does-not-exist.txt")
	s.Require().Error(err, "Running 'stat' on non-existent file should give an error")
}

func (s *DiskTestSuite) TestWorkingDirectory() {
	var fs filestore.FS

	fs = filestore.Disk("testdata")
	s.Require().Equal("testdata", fs.WorkingDirectory())

	fs = fs.ChangeDirectory("inner1")
	s.Require().Equal("testdata/inner1", fs.WorkingDirectory())

	fs = fs.ChangeDirectory("inner2")
	s.Require().Equal("testdata/inner1/inner2", fs.WorkingDirectory())

	fs = fs.ChangeDirectory("../..")
	s.Require().Equal("testdata", fs.WorkingDirectory())
}

func (s *DiskTestSuite) TestExists() {
	fs := filestore.Disk("testdata")

	s.Require().True(fs.Exists("."), "Current directory should exist")
	s.Require().True(fs.Exists(".."), "Parent directory should exist")

	s.Require().False(fs.Exists("testdata"), "Already in 'testdata' directory, so checking Exists('testdata') should be false.")

	// Real files/dirs should exist regardless of nesting
	s.Require().True(fs.Exists("hello.txt"), "Real file should exist")
	s.Require().True(fs.Exists("inner1"), "Real directory should exist")
	s.Require().True(fs.Exists("inner1/inner2/../foo.txt"), "Real file should exist when specifying relative path")
	s.Require().True(fs.Exists("inner1/inner2/.."), "Real dir should exist when specifying relative path")

	s.Require().False(fs.Exists("asldkfj"), "Non-existing entry should be false for Exists()")
	s.Require().False(fs.Exists("inner1/alskdjfalkdsfj.txt"), "Non-existing entry should be false for Exists() even when parent directory exists")

	// Bug fix where we weren't prepending the base directory to the path you provide.
	s.Require().True(fs.ChangeDirectory("inner1").Exists("inner2/../foo.txt"), "Real file should exist even after cd")
	s.Require().False(fs.ChangeDirectory("inner1").Exists("inner2/../nope.txt"), "Non-existing file should not exist even after cd")
}

func (s *DiskTestSuite) TestList_noFilters() {
	fs := filestore.Disk("testdata")

	files, err := fs.List("hello.txt")
	s.Require().Error(err, "File list for non-directories should return an error.")
	s.Require().Equal(0, len(files), "File list for non-directories should be empty.")

	files, err = fs.List(".")
	s.Require().NoError(err, "File list for current directory should not return an error.")
	s.Require().Equal(2, len(files), "File list for current directory should contain testdata/ items.")
	s.assertFile(files[0], "hello.txt")
	s.assertDir(files[1], "inner1")

	files, err = fs.List("inner1/inner2")
	s.Require().NoError(err, "File list for valid directories should not return an error.")
	s.Require().Equal(3, len(files))
	s.assertFile(files[0], "bar.txt")
	s.assertFile(files[1], "baz.log")
	s.assertFile(files[2], "blah.blah")
}

func (s *DiskTestSuite) TestList_withFilters() {
	fs := filestore.Disk("testdata")

	// Matches nothing
	files, err := fs.List(".", filestore.WithExt("fart"))
	s.Require().NoError(err, "File list filter with no results should not return an error.")
	s.Require().Equal(0, len(files), "File list with no results should be empty.")

	// One filter
	files, err = fs.List(".", filestore.WithExt("txt"))
	s.Require().NoError(err, "File list for current directory should not return an error.")
	s.Require().Equal(1, len(files), "File list for current directory should contain testdata/ '.txt.' items.")
	s.assertFile(files[0], "hello.txt")

	// Multiple filters
	files, err = fs.List("inner1/inner2", filestore.WithPattern("b*"), filestore.WithPattern("?a*"))
	s.Require().NoError(err, "File list with multiple filters should not return an error.")
	s.Require().Equal(2, len(files))
	s.assertFile(files[0], "bar.txt")
	s.assertFile(files[1], "baz.log")
}

// Removing a non-existent file should quietly do nothing.
func (s *DiskTestSuite) TestRemove_nonExistent() {
	err := filestore.Disk(s.tempDirPath).Remove("asldfjslkdfjasdf")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Removing non-existent file should NOT return an error")
	s.Require().Equal(6, len(files), "All 6 original entries (4 files, 2 dirs) should remain when removing non-existent file.")
}

func (s *DiskTestSuite) TestRemove_validFiles() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Remove("4.lebowski")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Removing valid file should NOT return an error")
	s.Require().Equal(5, len(files), "3 files and 2 dirs should remain when removing a valid file.")
	s.assertFile(files[0], "1.lebowski")
	s.assertFile(files[1], "2.lebowski")
	s.assertFile(files[2], "3.lebowski")
	s.assertDir(files[3], "dude")
	s.assertDir(files[4], "duderino")

	err = fs.Remove("2.lebowski")
	files = s.ls(s.tempDirPath)
	s.Require().NoError(err, "Removing valid file should NOT return an error")
	s.Require().Equal(4, len(files), "2 files and 2 dirs should remain when removing another valid file.")
	s.assertFile(files[0], "1.lebowski")
	s.assertFile(files[1], "3.lebowski")
	s.assertDir(files[2], "dude")
	s.assertDir(files[3], "duderino")

	err = fs.Remove("duderino/5.lebowski")
	files = s.ls(path.Join(s.tempDirPath, "duderino"))
	s.Require().NoError(err, "Removing valid file should NOT return an error")
	s.Require().Equal(1, len(files), "1 file should remain when removing one of the 2 files in duderino/")
	s.assertFile(files[0], "6.lebowski")
}

func (s *DiskTestSuite) TestRemove_validDirs() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Remove("dude")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Removing valid dir should NOT return an error")
	s.Require().Equal(5, len(files), "4 files and 1 dir should remain when removing a valid dir.")
	s.assertFile(files[0], "1.lebowski")
	s.assertFile(files[1], "2.lebowski")
	s.assertFile(files[2], "3.lebowski")
	s.assertFile(files[3], "4.lebowski")
	s.assertDir(files[4], "duderino")

	// The duderino dir has 2 files in it, so make sure that both are there before we delete
	// the directory and are gone after we delete it.
	files = s.ls(path.Join(s.tempDirPath, "duderino"))
	s.Require().Equal(2, len(files), "Should have 2 files in duderino/ directory before deleting it.")
	err = fs.Remove("duderino")
	files = s.ls(path.Join(s.tempDirPath, "duderino"))
	s.Require().NoError(err, "Removing valid dir should NOT return an error")
	s.Require().Equal(0, len(files), "Should have 0 files in duderino/ directory after deleting it.")
}

func (s *DiskTestSuite) assertFile(file filestore.FileInfo, name string) {
	s.Require().Equal(name, file.Name())
	s.Require().False(file.IsDir())
}

func (s *DiskTestSuite) assertDir(file filestore.FileInfo, name string) {
	s.Require().Equal(name, file.Name())
	s.Require().True(file.IsDir())
}

// Should overwrite existing file if you choose to move a file to a location for another existing file.
func (s *DiskTestSuite) TestMove_conflictFileToFile() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("1.lebowski", "2.lebowski")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Moving file to location that already has that name should NOT fail")
	s.Require().Equal(5, len(files), "Overwriting file should result in 1 less file in the directory.")
	s.assertFile(files[0], "2.lebowski")
	s.Require().Equal("jeff", s.read(s.tempDirPath, "2.lebowski"), "Moved file should overwrite 'walter' with 'jeff'.")
}

// Should not be able to rename a file to the name of an existing directory.
func (s *DiskTestSuite) TestMove_conflictFileToDir() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("1.lebowski", "dude")
	files := s.ls(s.tempDirPath)
	s.Require().Error(err, "Moving file to location of existing directory should fail.")
	s.Require().Equal(6, len(files), "Failed move should not change directory structure.")
}

// Should be able to use Move() to rename files
func (s *DiskTestSuite) TestMove_basicRename() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("1.lebowski", "jeff.lebowski")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Renaming file to unused name should not cause an error.")
	s.Require().Equal(6, len(files), "Renaming should not change number of files in directory.")
	s.assertFile(files[5], "jeff.lebowski")
	s.Require().Equal("jeff", s.read(s.tempDirPath, "jeff.lebowski"), "Moved file should contain original 'jeff' content.")
}

// Should be able to use Move() to move files from one directory to another.
func (s *DiskTestSuite) TestMove_fileToDirectory() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("1.lebowski", "dude/1.lebowski")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Moving file to another directory should not cause an error.")
	s.Require().Equal(5, len(files), "Original directory should have one less file in it.")
	s.Require().Equal("", s.read(s.tempDirPath, "1.lebowski"), "Original file location should not exist anymore.")
	s.Require().Equal("jeff", s.read(s.tempDirPath, "dude/1.lebowski"), "Moved file should contain original 'jeff' content.")
}

// Should be NOT able to rename a directory if another file with that name already exists.
func (s *DiskTestSuite) TestMove_directoryRenameConflictFile() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("duderino", "1.lebowski")
	files := s.ls(s.tempDirPath)
	s.Require().Error(err, "Renaming directory to the name of another file should fail.")
	s.Require().Equal(6, len(files), "Original parent directory should contain same number of files.")
}

// Should be NOT able to rename a directory if another directory with that name already exists.
func (s *DiskTestSuite) TestMove_directoryRenameConflictDir() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("duderino", "dude")
	files := s.ls(s.tempDirPath)
	s.Require().Error(err, "Renaming directory to the name of another directory should fail.")
	s.Require().Equal(6, len(files), "Original parent directory should contain same number of files.")
}

// Should be able to rename a directory.
func (s *DiskTestSuite) TestMove_directoryRename() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("duderino", "el duderino")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Renaming directory should not cause an error")
	s.Require().Equal(6, len(files), "Original parent directory should contain same number of files.")
	s.assertDir(files[5], "el duderino")

	// Make sure that the children of the renamed directory are still there.
	s.Require().Equal("jackie", s.read(s.tempDirPath, "el duderino/5.lebowski"), "Moved file should contain original content.")
	s.Require().Equal("nihilist", s.read(s.tempDirPath, "el duderino/6.lebowski"), "Moved file should contain original content.")
}

// Should be able to move a directory into another existing directory.
func (s *DiskTestSuite) TestMove_changeDirectoryStructure() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("duderino", "dude/duderino")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Moving directory inside another existing directory should not fail.")
	s.Require().Equal(5, len(files), "Original parent directory should have one less file in it.")
	files = s.ls(s.tempDirPath, "dude")
	s.Require().Equal(1, len(files), "New parent directory should contain the moved directory.")
	s.assertDir(files[0], "duderino")
}

// Moving file to location w/ non-existent path should create the path automatically.
func (s *DiskTestSuite) TestMove_autoCreateParentsForFile() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("1.lebowski", "dude/a/b/c/1.lebowski")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Moving file to location w/ non-existent path should not fail.")
	s.Require().Equal(5, len(files), "Original parent directory should have one less file in it.")
	s.assertFile(files[0], "2.lebowski")
	files = s.ls(s.tempDirPath, "dude/a/b/c")
	s.Require().Equal(1, len(files), "New parent directory should contain the moved file.")
	s.assertFile(files[0], "1.lebowski")
}

// Moving dir to location w/ non-existent path should create the path automatically.
func (s *DiskTestSuite) TestMove_autoCreateParentsForDir() {
	fs := filestore.Disk(s.tempDirPath)

	err := fs.Move("duderino", "dude/a/b/c/duderino")
	files := s.ls(s.tempDirPath)
	s.Require().NoError(err, "Moving dir to location w/ non-existent path should not fail.")
	s.Require().Equal(5, len(files), "Original parent directory should have one less file in it.")
	s.assertDir(files[4], "dude")
	files = s.ls(s.tempDirPath, "dude/a/b/c/duderino")
	s.Require().Equal(2, len(files), "Directory should contain all of its original files.")
	s.assertFile(files[0], "5.lebowski")
	s.assertFile(files[1], "6.lebowski")
}

func (s *DiskTestSuite) TestRead() {
	fs := filestore.Disk("testdata")

	// File that does not exist
	file, err := fs.Read("not-found.txt")
	s.Require().Error(err, "Reading invalid file in base directory should fail")
	file, err = fs.Read("a/b/c/not-found.txt")
	s.Require().Error(err, "Reading invalid file in base directory should fail")

	// File that DOES exist
	file, err = fs.Read("hello.txt")
	s.Require().NoError(err, "Reading valid file in base directory should not fail")
	s.Require().Equal("Hello World\n", s.toString(file))

	// File that DOES exist inside a child directory
	file, err = fs.Read("inner1/inner2/bar.txt")
	s.Require().NoError(err, "Reading valid file in base directory should not fail")
	s.Require().Equal("Bar\n", s.toString(file))

	// Trying to read a directory like it's a file should fail.
	_, err = fs.Read("inner1")
	s.Require().Error(err, "Reading directory as if it were a file should fail")
}

func (s *DiskTestSuite) TestWrite() {
	fs := filestore.Disk(s.tempDirPath)

	write := func(fileName string, content string) error {
		file, err := fs.Write(fileName)
		if err != nil {
			return err
		}
		_, _ = file.Write([]byte(content))
		return file.Close()
	}

	// Can overwrite an existing file.
	err := write("1.lebowski", "thank you donnie")
	s.Require().NoError(err, "Should be able to overwrite an existing file.")
	s.Require().Equal("thank you donnie", s.read(s.tempDirPath, "1.lebowski"), "Overwritten file should contain new data.")

	// Can write a brand-new file in the working directory
	err = write("x.lebowski", "abide")
	s.Require().NoError(err, "Should be able to write a new file in the source directory.")
	s.Require().Equal("abide", s.read(s.tempDirPath, "x.lebowski"), "Newly written file should contain proper data.")

	// Can write a brand-new file in an existing child directory
	err = write("dude/x.lebowski", "abide")
	s.Require().NoError(err, "Should be able to write a new file in a child directory.")
	s.Require().Equal("abide", s.read(s.tempDirPath, "dude/x.lebowski"), "Newly written file should contain proper data.")

	// Can write a brand-new file in a non-existing directory and have it auto-create the parents
	err = write("a/b/c/d/x.lebowski", "abide")
	s.Require().NoError(err, "Should be able to write a new file in a child directory.")
	s.Require().Equal("abide", s.read(s.tempDirPath, "a/b/c/d/x.lebowski"), "Newly written file should contain proper data.")
}

// Yes, our FS has a List() method, but this uses raw os.ReadDir() so that you can compare
// directory contents without relying on potentially broken implementations in our FS.
func (s *DiskTestSuite) ls(directorySegments ...string) []filestore.FileInfo {
	entries, _ := os.ReadDir(path.Join(directorySegments...))

	var infos []filestore.FileInfo
	for _, entry := range entries {
		file, _ := entry.Info()
		infos = append(infos, file)
	}
	return infos
}

func (s *DiskTestSuite) read(filePathSegments ...string) string {
	data, _ := os.ReadFile(path.Join(filePathSegments...))
	return string(data)
}

func (s *DiskTestSuite) toString(file filestore.ReaderFile) string {
	data, _ := io.ReadAll(file)
	return string(data)
}
