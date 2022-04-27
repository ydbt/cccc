package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:"version",
	Short:"服务版本",
	Long:"输出服务版本详情",
	Run:func(cmd *cobra.Command, args []string){
		fmt.Println(ms_version())
	},
}

func init(){
	rootCmd.AddCommand(versionCmd)
}


var (
	moduleType          string
	subSystemName       string
	subSystemVersion    string
	moduleName          string
	moduleFullName      string
	moduleVersion       string
	moduleRevision      string
	moduleShortRevision string
	vcsBranch           string
	vcsReposity         string
	commitTime          string
	buildJDK            string
	buildUser           string
	buildMachine        string
	buildTime           string
	customInfo          string
	developAuthor       string
	buildSDK            string
)

func ms_version() string {
	return fmt.Sprintf(`
***********************************************************************
Develop_Language: GOLANG
Module_Type: %s
Sub-system_Name: %s
Sub-system_Version: %s
Module_Name: %s
Module_Full_Name: %s
Module_Version: %s
Module_Revision: %s
Module_Short_Revision: %s
Vcs_Branch: %s
Vcs_Reposity: %s
Commit_Time: %s
Build Info: build by %s at %s on %s
Build_Jdk: %s
Custom_Info: %s
Develop_Author: %s
Build_Sdk: %s
Channelsoft(Beijing) Technologiges Co.,Ltd.
All rights reserved.
***********************************************************************
`,
		moduleType,
		subSystemName,
		subSystemVersion,
		moduleName,
		moduleFullName,
		moduleVersion,
		moduleRevision,
		moduleShortRevision,
		vcsBranch,
		vcsReposity,
		commitTime,
		buildUser,
		buildMachine,
		buildTime,
		buildJDK,
		customInfo,
		developAuthor,
		runtime.Version())
}

