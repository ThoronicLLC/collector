package cli

import "os"

func FileExists(filename string) bool {
  info, err := os.Stat(filename)
  if os.IsNotExist(err) {
    return false
  }
  return !info.IsDir()
}

func DirectoryExists(dir string) bool {
  info, err := os.Stat(dir)
  if os.IsNotExist(err) {
    return false
  }
  return info.IsDir()
}
