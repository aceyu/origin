package policy

import (
	"errors"
	"sort"

	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	kcmdutil "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

var deleteBindingRule = authorizationapi.PolicyRule{Verbs: util.NewStringSet("delete"), Resources: util.NewStringSet("rolebindings")}

type removeUserFromProjectOptions struct {
	bindingNamespace string
	client           client.Interface

	users []string
}

func NewCmdRemoveUserFromProject(f *clientcmd.Factory) *cobra.Command {
	options := &removeUserFromProjectOptions{}

	cmd := &cobra.Command{
		Use:   "remove-user  <user> [user]...",
		Short: "remove user from project",
		Long:  `remove user from project`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.complete(args); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			var err error
			if options.client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}
			if options.bindingNamespace, err = f.DefaultNamespace(); err != nil {
				kcmdutil.CheckErr(err)
			}
			if err := options.run(); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	return cmd
}

func (o *removeUserFromProjectOptions) complete(args []string) error {
	if len(args) < 1 {
		return errors.New("You must specify at least one argument: <user> [user]...")
	}

	o.users = args
	return nil
}

func (o *removeUserFromProjectOptions) run() error {
	bindingList, err := o.client.PolicyBindings(o.bindingNamespace).List(labels.Everything(), fields.Everything())
	if err != nil {
		return err
	}
	sort.Sort(authorizationapi.PolicyBindingSorter(bindingList.Items))

	for _, currPolicyBinding := range bindingList.Items {
		for _, currBinding := range authorizationapi.SortRoleBindings(currPolicyBinding.RoleBindings, true) {
			if !currBinding.Users.HasAny(o.users...) {
				continue
			}

			currBinding.Users.Delete(o.users...)

			_, err = o.client.RoleBindings(o.bindingNamespace).Update(&currBinding)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
