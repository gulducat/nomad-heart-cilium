package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	ciliumModels "github.com/cilium/cilium/api/v1/models"
	ciliumClient "github.com/cilium/cilium/pkg/client"
	cni "github.com/containernetworking/cni/libcni"
	cniSkel "github.com/containernetworking/cni/pkg/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	cniVer "github.com/containernetworking/cni/pkg/version"
)

var (
	// if not "IDE" env var, logging will go to pluginLog file,
	// and the stdout from the previous plugin in CNI chain (our stdin)
	// will be saved to saveStdin.
	isIDE     = os.Getenv("IDE") != ""
	pluginLog = "/tmp/cilium-heart-nomad.log"
	saveStdin = "/tmp/cilium-stdout"
)

type CNIArgs struct {
	cniTypes.CommonArgs

	NomadNamespace cniTypes.UnmarshallableString
	NomadTaskGroup cniTypes.UnmarshallableString
}

func main() {
	defer setupLogging()()

	cniSkel.PluginMain(
		cniFunc("add"),
		cniFunc("check"),
		cniFunc("delete"),
		cniVer.All,
		"label Cilium endpoints with metadata from Nomad",
	)
}

func cniFunc(cmd string) func(a *cniSkel.CmdArgs) error {
	return func(a *cniSkel.CmdArgs) error {
		writeFile(saveStdin, string(a.StdinData))
		log.Println("a.StdinData:", string(a.StdinData))

		// check out what cilium already did, and output its results.
		net, err := cni.ConfFromBytes(a.StdinData)
		if err != nil {
			return err
		}
		netConf := net.Network
		err = cniVer.ParsePrevResult(netConf)
		if err != nil {
			return err
		}
		if netConf.PrevResult == nil {
			errors.New("no PrevResult from a previous plugin in the CNI chain")
		}
		// write the result of cilium's hard work.
		err = netConf.PrevResult.Print()
		if err != nil {
			return err
		}

		// we only want to take special action on "add" commands,
		// otherwise stop here because cilium has done all the work.
		switch cmd {
		case "check", "delete":
			return nil
		}

		// now, tell cilium about some nomad stuff
		args := &CNIArgs{}
		err = cniTypes.LoadArgs(a.Args, args)
		if err != nil {
			return err
		}
		log.Printf("hiiii args: %+v", args)
		log.Printf("hiiii netconf: %+v", netConf)

		cc, err := ciliumClient.NewDefaultClient()
		if err != nil {
			return fmt.Errorf("failed to get cilium client: %w", err)
		}
		log.Printf("cilium: %+v", cc)
		eps, err := cc.EndpointList()
		if err != nil {
			return err
		}
		var ep *ciliumModels.Endpoint
		for _, e := range eps {
			if e.Status.ExternalIdentifiers.ContainerID == a.ContainerID {
				ep = e
				break
			}
		}
		if ep == nil {
			return errors.New("did not find cilium endpoint")
		}
		log.Printf("cilium endpoint: %d", ep.ID)
		id := strconv.Itoa(int(ep.ID))
		err = cc.EndpointLabelsPatch(id,
			// add
			ciliumModels.Labels{
				//fmt.Sprintf("nomad_alloc_id:%s", args.NomadAllocID),
				fmt.Sprintf("nomad_namespace:%s", args.NomadNamespace),
				fmt.Sprintf("nomad_taskgroup:%s", args.NomadTaskGroup),
			},
			nil)
		if err != nil {
			return err
		}
		err = cc.EndpointLabelsPatch(id, nil,
			// delete
			ciliumModels.Labels{"reserved:init"},
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func setupLogging() func() error {
	if isIDE {
		return nil
	}
	file, err := os.OpenFile(pluginLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("could not open log:", err)
	}
	log.Default().SetOutput(file)
	return file.Close
}

func writeFile(path, stdin string) {
	// for debugging, it can be nice to write cilium's stdout,
	// which we get here as stdin, to some file, which one can then
	// feed into the binary without needing Nomad involved.
	if isIDE {
		return
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("error opening chain file:", err)
	} else {
		if _, err = f.WriteString(stdin + "\n"); err != nil {
			log.Println("error writing file:", err)
		}
	}
}
