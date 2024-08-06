package spruce

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"

	"github.com/geofffranks/simpleyaml"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/starkandwayne/goutils/tree"
)

type mockedSSM struct {
	ssmiface.SSMAPI
	secretsmanageriface.SecretsManagerAPI

	MockGetParameter func(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

// AddTagsToResource implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).AddTagsToResource of mockedSSM.SSMAPI.
func (m *mockedSSM) AddTagsToResource(*ssm.AddTagsToResourceInput) (*ssm.AddTagsToResourceOutput, error) {
	panic("unimplemented")
}

// AddTagsToResourceRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).AddTagsToResourceRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) AddTagsToResourceRequest(*ssm.AddTagsToResourceInput) (*request.Request, *ssm.AddTagsToResourceOutput) {
	panic("unimplemented")
}

// AddTagsToResourceWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).AddTagsToResourceWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) AddTagsToResourceWithContext(context.Context, *ssm.AddTagsToResourceInput, ...request.Option) (*ssm.AddTagsToResourceOutput, error) {
	panic("unimplemented")
}

// AssociateOpsItemRelatedItem implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).AssociateOpsItemRelatedItem of mockedSSM.SSMAPI.
func (m *mockedSSM) AssociateOpsItemRelatedItem(*ssm.AssociateOpsItemRelatedItemInput) (*ssm.AssociateOpsItemRelatedItemOutput, error) {
	panic("unimplemented")
}

// AssociateOpsItemRelatedItemRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).AssociateOpsItemRelatedItemRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) AssociateOpsItemRelatedItemRequest(*ssm.AssociateOpsItemRelatedItemInput) (*request.Request, *ssm.AssociateOpsItemRelatedItemOutput) {
	panic("unimplemented")
}

// AssociateOpsItemRelatedItemWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).AssociateOpsItemRelatedItemWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) AssociateOpsItemRelatedItemWithContext(context.Context, *ssm.AssociateOpsItemRelatedItemInput, ...request.Option) (*ssm.AssociateOpsItemRelatedItemOutput, error) {
	panic("unimplemented")
}

// CancelCommand implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CancelCommand of mockedSSM.SSMAPI.
func (m *mockedSSM) CancelCommand(*ssm.CancelCommandInput) (*ssm.CancelCommandOutput, error) {
	panic("unimplemented")
}

// CancelCommandRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CancelCommandRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CancelCommandRequest(*ssm.CancelCommandInput) (*request.Request, *ssm.CancelCommandOutput) {
	panic("unimplemented")
}

// CancelCommandWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CancelCommandWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CancelCommandWithContext(context.Context, *ssm.CancelCommandInput, ...request.Option) (*ssm.CancelCommandOutput, error) {
	panic("unimplemented")
}

// CancelMaintenanceWindowExecution implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CancelMaintenanceWindowExecution of mockedSSM.SSMAPI.
func (m *mockedSSM) CancelMaintenanceWindowExecution(*ssm.CancelMaintenanceWindowExecutionInput) (*ssm.CancelMaintenanceWindowExecutionOutput, error) {
	panic("unimplemented")
}

// CancelMaintenanceWindowExecutionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CancelMaintenanceWindowExecutionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CancelMaintenanceWindowExecutionRequest(*ssm.CancelMaintenanceWindowExecutionInput) (*request.Request, *ssm.CancelMaintenanceWindowExecutionOutput) {
	panic("unimplemented")
}

// CancelMaintenanceWindowExecutionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CancelMaintenanceWindowExecutionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CancelMaintenanceWindowExecutionWithContext(context.Context, *ssm.CancelMaintenanceWindowExecutionInput, ...request.Option) (*ssm.CancelMaintenanceWindowExecutionOutput, error) {
	panic("unimplemented")
}

// CreateActivation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateActivation of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateActivation(*ssm.CreateActivationInput) (*ssm.CreateActivationOutput, error) {
	panic("unimplemented")
}

// CreateActivationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateActivationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateActivationRequest(*ssm.CreateActivationInput) (*request.Request, *ssm.CreateActivationOutput) {
	panic("unimplemented")
}

// CreateActivationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateActivationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateActivationWithContext(context.Context, *ssm.CreateActivationInput, ...request.Option) (*ssm.CreateActivationOutput, error) {
	panic("unimplemented")
}

// CreateAssociation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateAssociation of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateAssociation(*ssm.CreateAssociationInput) (*ssm.CreateAssociationOutput, error) {
	panic("unimplemented")
}

// CreateAssociationBatch implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateAssociationBatch of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateAssociationBatch(*ssm.CreateAssociationBatchInput) (*ssm.CreateAssociationBatchOutput, error) {
	panic("unimplemented")
}

// CreateAssociationBatchRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateAssociationBatchRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateAssociationBatchRequest(*ssm.CreateAssociationBatchInput) (*request.Request, *ssm.CreateAssociationBatchOutput) {
	panic("unimplemented")
}

// CreateAssociationBatchWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateAssociationBatchWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateAssociationBatchWithContext(context.Context, *ssm.CreateAssociationBatchInput, ...request.Option) (*ssm.CreateAssociationBatchOutput, error) {
	panic("unimplemented")
}

// CreateAssociationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateAssociationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateAssociationRequest(*ssm.CreateAssociationInput) (*request.Request, *ssm.CreateAssociationOutput) {
	panic("unimplemented")
}

// CreateAssociationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateAssociationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateAssociationWithContext(context.Context, *ssm.CreateAssociationInput, ...request.Option) (*ssm.CreateAssociationOutput, error) {
	panic("unimplemented")
}

// CreateDocument implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateDocument of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateDocument(*ssm.CreateDocumentInput) (*ssm.CreateDocumentOutput, error) {
	panic("unimplemented")
}

// CreateDocumentRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateDocumentRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateDocumentRequest(*ssm.CreateDocumentInput) (*request.Request, *ssm.CreateDocumentOutput) {
	panic("unimplemented")
}

// CreateDocumentWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateDocumentWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateDocumentWithContext(context.Context, *ssm.CreateDocumentInput, ...request.Option) (*ssm.CreateDocumentOutput, error) {
	panic("unimplemented")
}

// CreateMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateMaintenanceWindow(*ssm.CreateMaintenanceWindowInput) (*ssm.CreateMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// CreateMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateMaintenanceWindowRequest(*ssm.CreateMaintenanceWindowInput) (*request.Request, *ssm.CreateMaintenanceWindowOutput) {
	panic("unimplemented")
}

// CreateMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateMaintenanceWindowWithContext(context.Context, *ssm.CreateMaintenanceWindowInput, ...request.Option) (*ssm.CreateMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// CreateOpsItem implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateOpsItem of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateOpsItem(*ssm.CreateOpsItemInput) (*ssm.CreateOpsItemOutput, error) {
	panic("unimplemented")
}

// CreateOpsItemRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateOpsItemRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateOpsItemRequest(*ssm.CreateOpsItemInput) (*request.Request, *ssm.CreateOpsItemOutput) {
	panic("unimplemented")
}

// CreateOpsItemWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateOpsItemWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateOpsItemWithContext(context.Context, *ssm.CreateOpsItemInput, ...request.Option) (*ssm.CreateOpsItemOutput, error) {
	panic("unimplemented")
}

// CreateOpsMetadata implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateOpsMetadata of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateOpsMetadata(*ssm.CreateOpsMetadataInput) (*ssm.CreateOpsMetadataOutput, error) {
	panic("unimplemented")
}

// CreateOpsMetadataRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateOpsMetadataRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateOpsMetadataRequest(*ssm.CreateOpsMetadataInput) (*request.Request, *ssm.CreateOpsMetadataOutput) {
	panic("unimplemented")
}

// CreateOpsMetadataWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateOpsMetadataWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateOpsMetadataWithContext(context.Context, *ssm.CreateOpsMetadataInput, ...request.Option) (*ssm.CreateOpsMetadataOutput, error) {
	panic("unimplemented")
}

// CreatePatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreatePatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) CreatePatchBaseline(*ssm.CreatePatchBaselineInput) (*ssm.CreatePatchBaselineOutput, error) {
	panic("unimplemented")
}

// CreatePatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreatePatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreatePatchBaselineRequest(*ssm.CreatePatchBaselineInput) (*request.Request, *ssm.CreatePatchBaselineOutput) {
	panic("unimplemented")
}

// CreatePatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreatePatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreatePatchBaselineWithContext(context.Context, *ssm.CreatePatchBaselineInput, ...request.Option) (*ssm.CreatePatchBaselineOutput, error) {
	panic("unimplemented")
}

// CreateResourceDataSync implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateResourceDataSync of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateResourceDataSync(*ssm.CreateResourceDataSyncInput) (*ssm.CreateResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// CreateResourceDataSyncRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateResourceDataSyncRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateResourceDataSyncRequest(*ssm.CreateResourceDataSyncInput) (*request.Request, *ssm.CreateResourceDataSyncOutput) {
	panic("unimplemented")
}

// CreateResourceDataSyncWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).CreateResourceDataSyncWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) CreateResourceDataSyncWithContext(context.Context, *ssm.CreateResourceDataSyncInput, ...request.Option) (*ssm.CreateResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// DeleteActivation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteActivation of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteActivation(*ssm.DeleteActivationInput) (*ssm.DeleteActivationOutput, error) {
	panic("unimplemented")
}

// DeleteActivationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteActivationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteActivationRequest(*ssm.DeleteActivationInput) (*request.Request, *ssm.DeleteActivationOutput) {
	panic("unimplemented")
}

// DeleteActivationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteActivationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteActivationWithContext(context.Context, *ssm.DeleteActivationInput, ...request.Option) (*ssm.DeleteActivationOutput, error) {
	panic("unimplemented")
}

// DeleteAssociation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteAssociation of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteAssociation(*ssm.DeleteAssociationInput) (*ssm.DeleteAssociationOutput, error) {
	panic("unimplemented")
}

// DeleteAssociationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteAssociationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteAssociationRequest(*ssm.DeleteAssociationInput) (*request.Request, *ssm.DeleteAssociationOutput) {
	panic("unimplemented")
}

// DeleteAssociationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteAssociationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteAssociationWithContext(context.Context, *ssm.DeleteAssociationInput, ...request.Option) (*ssm.DeleteAssociationOutput, error) {
	panic("unimplemented")
}

// DeleteDocument implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteDocument of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteDocument(*ssm.DeleteDocumentInput) (*ssm.DeleteDocumentOutput, error) {
	panic("unimplemented")
}

// DeleteDocumentRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteDocumentRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteDocumentRequest(*ssm.DeleteDocumentInput) (*request.Request, *ssm.DeleteDocumentOutput) {
	panic("unimplemented")
}

// DeleteDocumentWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteDocumentWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteDocumentWithContext(context.Context, *ssm.DeleteDocumentInput, ...request.Option) (*ssm.DeleteDocumentOutput, error) {
	panic("unimplemented")
}

// DeleteInventory implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteInventory of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteInventory(*ssm.DeleteInventoryInput) (*ssm.DeleteInventoryOutput, error) {
	panic("unimplemented")
}

// DeleteInventoryRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteInventoryRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteInventoryRequest(*ssm.DeleteInventoryInput) (*request.Request, *ssm.DeleteInventoryOutput) {
	panic("unimplemented")
}

// DeleteInventoryWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteInventoryWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteInventoryWithContext(context.Context, *ssm.DeleteInventoryInput, ...request.Option) (*ssm.DeleteInventoryOutput, error) {
	panic("unimplemented")
}

// DeleteMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteMaintenanceWindow(*ssm.DeleteMaintenanceWindowInput) (*ssm.DeleteMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// DeleteMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteMaintenanceWindowRequest(*ssm.DeleteMaintenanceWindowInput) (*request.Request, *ssm.DeleteMaintenanceWindowOutput) {
	panic("unimplemented")
}

// DeleteMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteMaintenanceWindowWithContext(context.Context, *ssm.DeleteMaintenanceWindowInput, ...request.Option) (*ssm.DeleteMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// DeleteOpsMetadata implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteOpsMetadata of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteOpsMetadata(*ssm.DeleteOpsMetadataInput) (*ssm.DeleteOpsMetadataOutput, error) {
	panic("unimplemented")
}

// DeleteOpsMetadataRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteOpsMetadataRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteOpsMetadataRequest(*ssm.DeleteOpsMetadataInput) (*request.Request, *ssm.DeleteOpsMetadataOutput) {
	panic("unimplemented")
}

// DeleteOpsMetadataWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteOpsMetadataWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteOpsMetadataWithContext(context.Context, *ssm.DeleteOpsMetadataInput, ...request.Option) (*ssm.DeleteOpsMetadataOutput, error) {
	panic("unimplemented")
}

// DeleteParameter implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteParameter of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteParameter(*ssm.DeleteParameterInput) (*ssm.DeleteParameterOutput, error) {
	panic("unimplemented")
}

// DeleteParameterRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteParameterRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteParameterRequest(*ssm.DeleteParameterInput) (*request.Request, *ssm.DeleteParameterOutput) {
	panic("unimplemented")
}

// DeleteParameterWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteParameterWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteParameterWithContext(context.Context, *ssm.DeleteParameterInput, ...request.Option) (*ssm.DeleteParameterOutput, error) {
	panic("unimplemented")
}

// DeleteParameters implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteParameters of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteParameters(*ssm.DeleteParametersInput) (*ssm.DeleteParametersOutput, error) {
	panic("unimplemented")
}

// DeleteParametersRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteParametersRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteParametersRequest(*ssm.DeleteParametersInput) (*request.Request, *ssm.DeleteParametersOutput) {
	panic("unimplemented")
}

// DeleteParametersWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteParametersWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteParametersWithContext(context.Context, *ssm.DeleteParametersInput, ...request.Option) (*ssm.DeleteParametersOutput, error) {
	panic("unimplemented")
}

// DeletePatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeletePatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) DeletePatchBaseline(*ssm.DeletePatchBaselineInput) (*ssm.DeletePatchBaselineOutput, error) {
	panic("unimplemented")
}

// DeletePatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeletePatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeletePatchBaselineRequest(*ssm.DeletePatchBaselineInput) (*request.Request, *ssm.DeletePatchBaselineOutput) {
	panic("unimplemented")
}

// DeletePatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeletePatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeletePatchBaselineWithContext(context.Context, *ssm.DeletePatchBaselineInput, ...request.Option) (*ssm.DeletePatchBaselineOutput, error) {
	panic("unimplemented")
}

// DeleteResourceDataSync implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteResourceDataSync of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteResourceDataSync(*ssm.DeleteResourceDataSyncInput) (*ssm.DeleteResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// DeleteResourceDataSyncRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteResourceDataSyncRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteResourceDataSyncRequest(*ssm.DeleteResourceDataSyncInput) (*request.Request, *ssm.DeleteResourceDataSyncOutput) {
	panic("unimplemented")
}

// DeleteResourceDataSyncWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeleteResourceDataSyncWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeleteResourceDataSyncWithContext(context.Context, *ssm.DeleteResourceDataSyncInput, ...request.Option) (*ssm.DeleteResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// DeleteResourcePolicy implements ssmiface.SSMAPI.
func (m *mockedSSM) DeleteResourcePolicy(*ssm.DeleteResourcePolicyInput) (*ssm.DeleteResourcePolicyOutput, error) {
	panic("unimplemented")
}

// DeleteResourcePolicyRequest implements ssmiface.SSMAPI.
func (m *mockedSSM) DeleteResourcePolicyRequest(*ssm.DeleteResourcePolicyInput) (*request.Request, *ssm.DeleteResourcePolicyOutput) {
	panic("unimplemented")
}

// DeleteResourcePolicyWithContext implements ssmiface.SSMAPI.
func (m *mockedSSM) DeleteResourcePolicyWithContext(context.Context, *ssm.DeleteResourcePolicyInput, ...request.Option) (*ssm.DeleteResourcePolicyOutput, error) {
	panic("unimplemented")
}

// DeregisterManagedInstance implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterManagedInstance of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterManagedInstance(*ssm.DeregisterManagedInstanceInput) (*ssm.DeregisterManagedInstanceOutput, error) {
	panic("unimplemented")
}

// DeregisterManagedInstanceRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterManagedInstanceRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterManagedInstanceRequest(*ssm.DeregisterManagedInstanceInput) (*request.Request, *ssm.DeregisterManagedInstanceOutput) {
	panic("unimplemented")
}

// DeregisterManagedInstanceWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterManagedInstanceWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterManagedInstanceWithContext(context.Context, *ssm.DeregisterManagedInstanceInput, ...request.Option) (*ssm.DeregisterManagedInstanceOutput, error) {
	panic("unimplemented")
}

// DeregisterPatchBaselineForPatchGroup implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterPatchBaselineForPatchGroup of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterPatchBaselineForPatchGroup(*ssm.DeregisterPatchBaselineForPatchGroupInput) (*ssm.DeregisterPatchBaselineForPatchGroupOutput, error) {
	panic("unimplemented")
}

// DeregisterPatchBaselineForPatchGroupRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterPatchBaselineForPatchGroupRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterPatchBaselineForPatchGroupRequest(*ssm.DeregisterPatchBaselineForPatchGroupInput) (*request.Request, *ssm.DeregisterPatchBaselineForPatchGroupOutput) {
	panic("unimplemented")
}

// DeregisterPatchBaselineForPatchGroupWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterPatchBaselineForPatchGroupWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterPatchBaselineForPatchGroupWithContext(context.Context, *ssm.DeregisterPatchBaselineForPatchGroupInput, ...request.Option) (*ssm.DeregisterPatchBaselineForPatchGroupOutput, error) {
	panic("unimplemented")
}

// DeregisterTargetFromMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterTargetFromMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterTargetFromMaintenanceWindow(*ssm.DeregisterTargetFromMaintenanceWindowInput) (*ssm.DeregisterTargetFromMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// DeregisterTargetFromMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterTargetFromMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterTargetFromMaintenanceWindowRequest(*ssm.DeregisterTargetFromMaintenanceWindowInput) (*request.Request, *ssm.DeregisterTargetFromMaintenanceWindowOutput) {
	panic("unimplemented")
}

// DeregisterTargetFromMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterTargetFromMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterTargetFromMaintenanceWindowWithContext(context.Context, *ssm.DeregisterTargetFromMaintenanceWindowInput, ...request.Option) (*ssm.DeregisterTargetFromMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// DeregisterTaskFromMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterTaskFromMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterTaskFromMaintenanceWindow(*ssm.DeregisterTaskFromMaintenanceWindowInput) (*ssm.DeregisterTaskFromMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// DeregisterTaskFromMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterTaskFromMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterTaskFromMaintenanceWindowRequest(*ssm.DeregisterTaskFromMaintenanceWindowInput) (*request.Request, *ssm.DeregisterTaskFromMaintenanceWindowOutput) {
	panic("unimplemented")
}

// DeregisterTaskFromMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DeregisterTaskFromMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DeregisterTaskFromMaintenanceWindowWithContext(context.Context, *ssm.DeregisterTaskFromMaintenanceWindowInput, ...request.Option) (*ssm.DeregisterTaskFromMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// DescribeActivations implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeActivations of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeActivations(*ssm.DescribeActivationsInput) (*ssm.DescribeActivationsOutput, error) {
	panic("unimplemented")
}

// DescribeActivationsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeActivationsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeActivationsPages(*ssm.DescribeActivationsInput, func(*ssm.DescribeActivationsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeActivationsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeActivationsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeActivationsPagesWithContext(context.Context, *ssm.DescribeActivationsInput, func(*ssm.DescribeActivationsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeActivationsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeActivationsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeActivationsRequest(*ssm.DescribeActivationsInput) (*request.Request, *ssm.DescribeActivationsOutput) {
	panic("unimplemented")
}

// DescribeActivationsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeActivationsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeActivationsWithContext(context.Context, *ssm.DescribeActivationsInput, ...request.Option) (*ssm.DescribeActivationsOutput, error) {
	panic("unimplemented")
}

// DescribeAssociation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociation of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociation(*ssm.DescribeAssociationInput) (*ssm.DescribeAssociationOutput, error) {
	panic("unimplemented")
}

// DescribeAssociationExecutionTargets implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionTargets of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionTargets(*ssm.DescribeAssociationExecutionTargetsInput) (*ssm.DescribeAssociationExecutionTargetsOutput, error) {
	panic("unimplemented")
}

// DescribeAssociationExecutionTargetsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionTargetsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionTargetsPages(*ssm.DescribeAssociationExecutionTargetsInput, func(*ssm.DescribeAssociationExecutionTargetsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeAssociationExecutionTargetsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionTargetsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionTargetsPagesWithContext(context.Context, *ssm.DescribeAssociationExecutionTargetsInput, func(*ssm.DescribeAssociationExecutionTargetsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeAssociationExecutionTargetsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionTargetsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionTargetsRequest(*ssm.DescribeAssociationExecutionTargetsInput) (*request.Request, *ssm.DescribeAssociationExecutionTargetsOutput) {
	panic("unimplemented")
}

// DescribeAssociationExecutionTargetsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionTargetsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionTargetsWithContext(context.Context, *ssm.DescribeAssociationExecutionTargetsInput, ...request.Option) (*ssm.DescribeAssociationExecutionTargetsOutput, error) {
	panic("unimplemented")
}

// DescribeAssociationExecutions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutions of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutions(*ssm.DescribeAssociationExecutionsInput) (*ssm.DescribeAssociationExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeAssociationExecutionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionsPages(*ssm.DescribeAssociationExecutionsInput, func(*ssm.DescribeAssociationExecutionsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeAssociationExecutionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionsPagesWithContext(context.Context, *ssm.DescribeAssociationExecutionsInput, func(*ssm.DescribeAssociationExecutionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeAssociationExecutionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionsRequest(*ssm.DescribeAssociationExecutionsInput) (*request.Request, *ssm.DescribeAssociationExecutionsOutput) {
	panic("unimplemented")
}

// DescribeAssociationExecutionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationExecutionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationExecutionsWithContext(context.Context, *ssm.DescribeAssociationExecutionsInput, ...request.Option) (*ssm.DescribeAssociationExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeAssociationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationRequest(*ssm.DescribeAssociationInput) (*request.Request, *ssm.DescribeAssociationOutput) {
	panic("unimplemented")
}

// DescribeAssociationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAssociationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAssociationWithContext(context.Context, *ssm.DescribeAssociationInput, ...request.Option) (*ssm.DescribeAssociationOutput, error) {
	panic("unimplemented")
}

// DescribeAutomationExecutions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationExecutions of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationExecutions(*ssm.DescribeAutomationExecutionsInput) (*ssm.DescribeAutomationExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeAutomationExecutionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationExecutionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationExecutionsPages(*ssm.DescribeAutomationExecutionsInput, func(*ssm.DescribeAutomationExecutionsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeAutomationExecutionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationExecutionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationExecutionsPagesWithContext(context.Context, *ssm.DescribeAutomationExecutionsInput, func(*ssm.DescribeAutomationExecutionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeAutomationExecutionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationExecutionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationExecutionsRequest(*ssm.DescribeAutomationExecutionsInput) (*request.Request, *ssm.DescribeAutomationExecutionsOutput) {
	panic("unimplemented")
}

// DescribeAutomationExecutionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationExecutionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationExecutionsWithContext(context.Context, *ssm.DescribeAutomationExecutionsInput, ...request.Option) (*ssm.DescribeAutomationExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeAutomationStepExecutions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationStepExecutions of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationStepExecutions(*ssm.DescribeAutomationStepExecutionsInput) (*ssm.DescribeAutomationStepExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeAutomationStepExecutionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationStepExecutionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationStepExecutionsPages(*ssm.DescribeAutomationStepExecutionsInput, func(*ssm.DescribeAutomationStepExecutionsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeAutomationStepExecutionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationStepExecutionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationStepExecutionsPagesWithContext(context.Context, *ssm.DescribeAutomationStepExecutionsInput, func(*ssm.DescribeAutomationStepExecutionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeAutomationStepExecutionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationStepExecutionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationStepExecutionsRequest(*ssm.DescribeAutomationStepExecutionsInput) (*request.Request, *ssm.DescribeAutomationStepExecutionsOutput) {
	panic("unimplemented")
}

// DescribeAutomationStepExecutionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAutomationStepExecutionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAutomationStepExecutionsWithContext(context.Context, *ssm.DescribeAutomationStepExecutionsInput, ...request.Option) (*ssm.DescribeAutomationStepExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeAvailablePatches implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAvailablePatches of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAvailablePatches(*ssm.DescribeAvailablePatchesInput) (*ssm.DescribeAvailablePatchesOutput, error) {
	panic("unimplemented")
}

// DescribeAvailablePatchesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAvailablePatchesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAvailablePatchesPages(*ssm.DescribeAvailablePatchesInput, func(*ssm.DescribeAvailablePatchesOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeAvailablePatchesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAvailablePatchesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAvailablePatchesPagesWithContext(context.Context, *ssm.DescribeAvailablePatchesInput, func(*ssm.DescribeAvailablePatchesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeAvailablePatchesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAvailablePatchesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAvailablePatchesRequest(*ssm.DescribeAvailablePatchesInput) (*request.Request, *ssm.DescribeAvailablePatchesOutput) {
	panic("unimplemented")
}

// DescribeAvailablePatchesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeAvailablePatchesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeAvailablePatchesWithContext(context.Context, *ssm.DescribeAvailablePatchesInput, ...request.Option) (*ssm.DescribeAvailablePatchesOutput, error) {
	panic("unimplemented")
}

// DescribeDocument implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeDocument of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeDocument(*ssm.DescribeDocumentInput) (*ssm.DescribeDocumentOutput, error) {
	panic("unimplemented")
}

// DescribeDocumentPermission implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeDocumentPermission of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeDocumentPermission(*ssm.DescribeDocumentPermissionInput) (*ssm.DescribeDocumentPermissionOutput, error) {
	panic("unimplemented")
}

// DescribeDocumentPermissionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeDocumentPermissionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeDocumentPermissionRequest(*ssm.DescribeDocumentPermissionInput) (*request.Request, *ssm.DescribeDocumentPermissionOutput) {
	panic("unimplemented")
}

// DescribeDocumentPermissionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeDocumentPermissionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeDocumentPermissionWithContext(context.Context, *ssm.DescribeDocumentPermissionInput, ...request.Option) (*ssm.DescribeDocumentPermissionOutput, error) {
	panic("unimplemented")
}

// DescribeDocumentRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeDocumentRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeDocumentRequest(*ssm.DescribeDocumentInput) (*request.Request, *ssm.DescribeDocumentOutput) {
	panic("unimplemented")
}

// DescribeDocumentWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeDocumentWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeDocumentWithContext(context.Context, *ssm.DescribeDocumentInput, ...request.Option) (*ssm.DescribeDocumentOutput, error) {
	panic("unimplemented")
}

// DescribeEffectiveInstanceAssociations implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectiveInstanceAssociations of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectiveInstanceAssociations(*ssm.DescribeEffectiveInstanceAssociationsInput) (*ssm.DescribeEffectiveInstanceAssociationsOutput, error) {
	panic("unimplemented")
}

// DescribeEffectiveInstanceAssociationsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectiveInstanceAssociationsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectiveInstanceAssociationsPages(*ssm.DescribeEffectiveInstanceAssociationsInput, func(*ssm.DescribeEffectiveInstanceAssociationsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeEffectiveInstanceAssociationsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectiveInstanceAssociationsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectiveInstanceAssociationsPagesWithContext(context.Context, *ssm.DescribeEffectiveInstanceAssociationsInput, func(*ssm.DescribeEffectiveInstanceAssociationsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeEffectiveInstanceAssociationsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectiveInstanceAssociationsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectiveInstanceAssociationsRequest(*ssm.DescribeEffectiveInstanceAssociationsInput) (*request.Request, *ssm.DescribeEffectiveInstanceAssociationsOutput) {
	panic("unimplemented")
}

// DescribeEffectiveInstanceAssociationsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectiveInstanceAssociationsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectiveInstanceAssociationsWithContext(context.Context, *ssm.DescribeEffectiveInstanceAssociationsInput, ...request.Option) (*ssm.DescribeEffectiveInstanceAssociationsOutput, error) {
	panic("unimplemented")
}

// DescribeEffectivePatchesForPatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectivePatchesForPatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectivePatchesForPatchBaseline(*ssm.DescribeEffectivePatchesForPatchBaselineInput) (*ssm.DescribeEffectivePatchesForPatchBaselineOutput, error) {
	panic("unimplemented")
}

// DescribeEffectivePatchesForPatchBaselinePages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectivePatchesForPatchBaselinePages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectivePatchesForPatchBaselinePages(*ssm.DescribeEffectivePatchesForPatchBaselineInput, func(*ssm.DescribeEffectivePatchesForPatchBaselineOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeEffectivePatchesForPatchBaselinePagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectivePatchesForPatchBaselinePagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectivePatchesForPatchBaselinePagesWithContext(context.Context, *ssm.DescribeEffectivePatchesForPatchBaselineInput, func(*ssm.DescribeEffectivePatchesForPatchBaselineOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeEffectivePatchesForPatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectivePatchesForPatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectivePatchesForPatchBaselineRequest(*ssm.DescribeEffectivePatchesForPatchBaselineInput) (*request.Request, *ssm.DescribeEffectivePatchesForPatchBaselineOutput) {
	panic("unimplemented")
}

// DescribeEffectivePatchesForPatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeEffectivePatchesForPatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeEffectivePatchesForPatchBaselineWithContext(context.Context, *ssm.DescribeEffectivePatchesForPatchBaselineInput, ...request.Option) (*ssm.DescribeEffectivePatchesForPatchBaselineOutput, error) {
	panic("unimplemented")
}

// DescribeInstanceAssociationsStatus implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceAssociationsStatus of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceAssociationsStatus(*ssm.DescribeInstanceAssociationsStatusInput) (*ssm.DescribeInstanceAssociationsStatusOutput, error) {
	panic("unimplemented")
}

// DescribeInstanceAssociationsStatusPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceAssociationsStatusPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceAssociationsStatusPages(*ssm.DescribeInstanceAssociationsStatusInput, func(*ssm.DescribeInstanceAssociationsStatusOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeInstanceAssociationsStatusPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceAssociationsStatusPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceAssociationsStatusPagesWithContext(context.Context, *ssm.DescribeInstanceAssociationsStatusInput, func(*ssm.DescribeInstanceAssociationsStatusOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeInstanceAssociationsStatusRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceAssociationsStatusRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceAssociationsStatusRequest(*ssm.DescribeInstanceAssociationsStatusInput) (*request.Request, *ssm.DescribeInstanceAssociationsStatusOutput) {
	panic("unimplemented")
}

// DescribeInstanceAssociationsStatusWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceAssociationsStatusWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceAssociationsStatusWithContext(context.Context, *ssm.DescribeInstanceAssociationsStatusInput, ...request.Option) (*ssm.DescribeInstanceAssociationsStatusOutput, error) {
	panic("unimplemented")
}

// DescribeInstanceInformation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceInformation of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceInformation(*ssm.DescribeInstanceInformationInput) (*ssm.DescribeInstanceInformationOutput, error) {
	panic("unimplemented")
}

// DescribeInstanceInformationPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceInformationPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceInformationPages(*ssm.DescribeInstanceInformationInput, func(*ssm.DescribeInstanceInformationOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeInstanceInformationPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceInformationPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceInformationPagesWithContext(context.Context, *ssm.DescribeInstanceInformationInput, func(*ssm.DescribeInstanceInformationOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeInstanceInformationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceInformationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceInformationRequest(*ssm.DescribeInstanceInformationInput) (*request.Request, *ssm.DescribeInstanceInformationOutput) {
	panic("unimplemented")
}

// DescribeInstanceInformationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstanceInformationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstanceInformationWithContext(context.Context, *ssm.DescribeInstanceInformationInput, ...request.Option) (*ssm.DescribeInstanceInformationOutput, error) {
	panic("unimplemented")
}

// DescribeInstancePatchStates implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStates of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStates(*ssm.DescribeInstancePatchStatesInput) (*ssm.DescribeInstancePatchStatesOutput, error) {
	panic("unimplemented")
}

// DescribeInstancePatchStatesForPatchGroup implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesForPatchGroup of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesForPatchGroup(*ssm.DescribeInstancePatchStatesForPatchGroupInput) (*ssm.DescribeInstancePatchStatesForPatchGroupOutput, error) {
	panic("unimplemented")
}

// DescribeInstancePatchStatesForPatchGroupPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesForPatchGroupPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesForPatchGroupPages(*ssm.DescribeInstancePatchStatesForPatchGroupInput, func(*ssm.DescribeInstancePatchStatesForPatchGroupOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeInstancePatchStatesForPatchGroupPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesForPatchGroupPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesForPatchGroupPagesWithContext(context.Context, *ssm.DescribeInstancePatchStatesForPatchGroupInput, func(*ssm.DescribeInstancePatchStatesForPatchGroupOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeInstancePatchStatesForPatchGroupRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesForPatchGroupRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesForPatchGroupRequest(*ssm.DescribeInstancePatchStatesForPatchGroupInput) (*request.Request, *ssm.DescribeInstancePatchStatesForPatchGroupOutput) {
	panic("unimplemented")
}

// DescribeInstancePatchStatesForPatchGroupWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesForPatchGroupWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesForPatchGroupWithContext(context.Context, *ssm.DescribeInstancePatchStatesForPatchGroupInput, ...request.Option) (*ssm.DescribeInstancePatchStatesForPatchGroupOutput, error) {
	panic("unimplemented")
}

// DescribeInstancePatchStatesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesPages(*ssm.DescribeInstancePatchStatesInput, func(*ssm.DescribeInstancePatchStatesOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeInstancePatchStatesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesPagesWithContext(context.Context, *ssm.DescribeInstancePatchStatesInput, func(*ssm.DescribeInstancePatchStatesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeInstancePatchStatesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesRequest(*ssm.DescribeInstancePatchStatesInput) (*request.Request, *ssm.DescribeInstancePatchStatesOutput) {
	panic("unimplemented")
}

// DescribeInstancePatchStatesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchStatesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchStatesWithContext(context.Context, *ssm.DescribeInstancePatchStatesInput, ...request.Option) (*ssm.DescribeInstancePatchStatesOutput, error) {
	panic("unimplemented")
}

// DescribeInstancePatches implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatches of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatches(*ssm.DescribeInstancePatchesInput) (*ssm.DescribeInstancePatchesOutput, error) {
	panic("unimplemented")
}

// DescribeInstancePatchesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchesPages(*ssm.DescribeInstancePatchesInput, func(*ssm.DescribeInstancePatchesOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeInstancePatchesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchesPagesWithContext(context.Context, *ssm.DescribeInstancePatchesInput, func(*ssm.DescribeInstancePatchesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeInstancePatchesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchesRequest(*ssm.DescribeInstancePatchesInput) (*request.Request, *ssm.DescribeInstancePatchesOutput) {
	panic("unimplemented")
}

// DescribeInstancePatchesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInstancePatchesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInstancePatchesWithContext(context.Context, *ssm.DescribeInstancePatchesInput, ...request.Option) (*ssm.DescribeInstancePatchesOutput, error) {
	panic("unimplemented")
}

// DescribeInventoryDeletions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInventoryDeletions of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInventoryDeletions(*ssm.DescribeInventoryDeletionsInput) (*ssm.DescribeInventoryDeletionsOutput, error) {
	panic("unimplemented")
}

// DescribeInventoryDeletionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInventoryDeletionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInventoryDeletionsPages(*ssm.DescribeInventoryDeletionsInput, func(*ssm.DescribeInventoryDeletionsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeInventoryDeletionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInventoryDeletionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInventoryDeletionsPagesWithContext(context.Context, *ssm.DescribeInventoryDeletionsInput, func(*ssm.DescribeInventoryDeletionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeInventoryDeletionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInventoryDeletionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInventoryDeletionsRequest(*ssm.DescribeInventoryDeletionsInput) (*request.Request, *ssm.DescribeInventoryDeletionsOutput) {
	panic("unimplemented")
}

// DescribeInventoryDeletionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeInventoryDeletionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeInventoryDeletionsWithContext(context.Context, *ssm.DescribeInventoryDeletionsInput, ...request.Option) (*ssm.DescribeInventoryDeletionsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTaskInvocations implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTaskInvocations of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTaskInvocations(*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsInput) (*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTaskInvocationsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTaskInvocationsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTaskInvocationsPages(*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsInput, func(*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTaskInvocationsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTaskInvocationsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTaskInvocationsPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowExecutionTaskInvocationsInput, func(*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTaskInvocationsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTaskInvocationsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTaskInvocationsRequest(*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsInput) (*request.Request, *ssm.DescribeMaintenanceWindowExecutionTaskInvocationsOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTaskInvocationsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTaskInvocationsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTaskInvocationsWithContext(context.Context, *ssm.DescribeMaintenanceWindowExecutionTaskInvocationsInput, ...request.Option) (*ssm.DescribeMaintenanceWindowExecutionTaskInvocationsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTasks implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTasks of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTasks(*ssm.DescribeMaintenanceWindowExecutionTasksInput) (*ssm.DescribeMaintenanceWindowExecutionTasksOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTasksPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTasksPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTasksPages(*ssm.DescribeMaintenanceWindowExecutionTasksInput, func(*ssm.DescribeMaintenanceWindowExecutionTasksOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTasksPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTasksPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTasksPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowExecutionTasksInput, func(*ssm.DescribeMaintenanceWindowExecutionTasksOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTasksRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTasksRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTasksRequest(*ssm.DescribeMaintenanceWindowExecutionTasksInput) (*request.Request, *ssm.DescribeMaintenanceWindowExecutionTasksOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionTasksWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionTasksWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionTasksWithContext(context.Context, *ssm.DescribeMaintenanceWindowExecutionTasksInput, ...request.Option) (*ssm.DescribeMaintenanceWindowExecutionTasksOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutions of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutions(*ssm.DescribeMaintenanceWindowExecutionsInput) (*ssm.DescribeMaintenanceWindowExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionsPages(*ssm.DescribeMaintenanceWindowExecutionsInput, func(*ssm.DescribeMaintenanceWindowExecutionsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionsPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowExecutionsInput, func(*ssm.DescribeMaintenanceWindowExecutionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionsRequest(*ssm.DescribeMaintenanceWindowExecutionsInput) (*request.Request, *ssm.DescribeMaintenanceWindowExecutionsOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowExecutionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowExecutionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowExecutionsWithContext(context.Context, *ssm.DescribeMaintenanceWindowExecutionsInput, ...request.Option) (*ssm.DescribeMaintenanceWindowExecutionsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowSchedule implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowSchedule of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowSchedule(*ssm.DescribeMaintenanceWindowScheduleInput) (*ssm.DescribeMaintenanceWindowScheduleOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowSchedulePages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowSchedulePages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowSchedulePages(*ssm.DescribeMaintenanceWindowScheduleInput, func(*ssm.DescribeMaintenanceWindowScheduleOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowSchedulePagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowSchedulePagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowSchedulePagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowScheduleInput, func(*ssm.DescribeMaintenanceWindowScheduleOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowScheduleRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowScheduleRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowScheduleRequest(*ssm.DescribeMaintenanceWindowScheduleInput) (*request.Request, *ssm.DescribeMaintenanceWindowScheduleOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowScheduleWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowScheduleWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowScheduleWithContext(context.Context, *ssm.DescribeMaintenanceWindowScheduleInput, ...request.Option) (*ssm.DescribeMaintenanceWindowScheduleOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTargets implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTargets of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTargets(*ssm.DescribeMaintenanceWindowTargetsInput) (*ssm.DescribeMaintenanceWindowTargetsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTargetsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTargetsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTargetsPages(*ssm.DescribeMaintenanceWindowTargetsInput, func(*ssm.DescribeMaintenanceWindowTargetsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTargetsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTargetsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTargetsPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowTargetsInput, func(*ssm.DescribeMaintenanceWindowTargetsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTargetsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTargetsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTargetsRequest(*ssm.DescribeMaintenanceWindowTargetsInput) (*request.Request, *ssm.DescribeMaintenanceWindowTargetsOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTargetsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTargetsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTargetsWithContext(context.Context, *ssm.DescribeMaintenanceWindowTargetsInput, ...request.Option) (*ssm.DescribeMaintenanceWindowTargetsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTasks implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTasks of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTasks(*ssm.DescribeMaintenanceWindowTasksInput) (*ssm.DescribeMaintenanceWindowTasksOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTasksPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTasksPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTasksPages(*ssm.DescribeMaintenanceWindowTasksInput, func(*ssm.DescribeMaintenanceWindowTasksOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTasksPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTasksPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTasksPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowTasksInput, func(*ssm.DescribeMaintenanceWindowTasksOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTasksRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTasksRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTasksRequest(*ssm.DescribeMaintenanceWindowTasksInput) (*request.Request, *ssm.DescribeMaintenanceWindowTasksOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowTasksWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowTasksWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowTasksWithContext(context.Context, *ssm.DescribeMaintenanceWindowTasksInput, ...request.Option) (*ssm.DescribeMaintenanceWindowTasksOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindows implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindows of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindows(*ssm.DescribeMaintenanceWindowsInput) (*ssm.DescribeMaintenanceWindowsOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsForTarget implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsForTarget of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsForTarget(*ssm.DescribeMaintenanceWindowsForTargetInput) (*ssm.DescribeMaintenanceWindowsForTargetOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsForTargetPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsForTargetPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsForTargetPages(*ssm.DescribeMaintenanceWindowsForTargetInput, func(*ssm.DescribeMaintenanceWindowsForTargetOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsForTargetPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsForTargetPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsForTargetPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowsForTargetInput, func(*ssm.DescribeMaintenanceWindowsForTargetOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsForTargetRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsForTargetRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsForTargetRequest(*ssm.DescribeMaintenanceWindowsForTargetInput) (*request.Request, *ssm.DescribeMaintenanceWindowsForTargetOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsForTargetWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsForTargetWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsForTargetWithContext(context.Context, *ssm.DescribeMaintenanceWindowsForTargetInput, ...request.Option) (*ssm.DescribeMaintenanceWindowsForTargetOutput, error) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsPages(*ssm.DescribeMaintenanceWindowsInput, func(*ssm.DescribeMaintenanceWindowsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsPagesWithContext(context.Context, *ssm.DescribeMaintenanceWindowsInput, func(*ssm.DescribeMaintenanceWindowsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsRequest(*ssm.DescribeMaintenanceWindowsInput) (*request.Request, *ssm.DescribeMaintenanceWindowsOutput) {
	panic("unimplemented")
}

// DescribeMaintenanceWindowsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeMaintenanceWindowsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeMaintenanceWindowsWithContext(context.Context, *ssm.DescribeMaintenanceWindowsInput, ...request.Option) (*ssm.DescribeMaintenanceWindowsOutput, error) {
	panic("unimplemented")
}

// DescribeOpsItems implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeOpsItems of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeOpsItems(*ssm.DescribeOpsItemsInput) (*ssm.DescribeOpsItemsOutput, error) {
	panic("unimplemented")
}

// DescribeOpsItemsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeOpsItemsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeOpsItemsPages(*ssm.DescribeOpsItemsInput, func(*ssm.DescribeOpsItemsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeOpsItemsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeOpsItemsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeOpsItemsPagesWithContext(context.Context, *ssm.DescribeOpsItemsInput, func(*ssm.DescribeOpsItemsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeOpsItemsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeOpsItemsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeOpsItemsRequest(*ssm.DescribeOpsItemsInput) (*request.Request, *ssm.DescribeOpsItemsOutput) {
	panic("unimplemented")
}

// DescribeOpsItemsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeOpsItemsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeOpsItemsWithContext(context.Context, *ssm.DescribeOpsItemsInput, ...request.Option) (*ssm.DescribeOpsItemsOutput, error) {
	panic("unimplemented")
}

// DescribeParameters implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeParameters of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeParameters(*ssm.DescribeParametersInput) (*ssm.DescribeParametersOutput, error) {
	panic("unimplemented")
}

// DescribeParametersPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeParametersPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeParametersPages(*ssm.DescribeParametersInput, func(*ssm.DescribeParametersOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeParametersPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeParametersPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeParametersPagesWithContext(context.Context, *ssm.DescribeParametersInput, func(*ssm.DescribeParametersOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeParametersRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeParametersRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeParametersRequest(*ssm.DescribeParametersInput) (*request.Request, *ssm.DescribeParametersOutput) {
	panic("unimplemented")
}

// DescribeParametersWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeParametersWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeParametersWithContext(context.Context, *ssm.DescribeParametersInput, ...request.Option) (*ssm.DescribeParametersOutput, error) {
	panic("unimplemented")
}

// DescribePatchBaselines implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchBaselines of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchBaselines(*ssm.DescribePatchBaselinesInput) (*ssm.DescribePatchBaselinesOutput, error) {
	panic("unimplemented")
}

// DescribePatchBaselinesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchBaselinesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchBaselinesPages(*ssm.DescribePatchBaselinesInput, func(*ssm.DescribePatchBaselinesOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribePatchBaselinesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchBaselinesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchBaselinesPagesWithContext(context.Context, *ssm.DescribePatchBaselinesInput, func(*ssm.DescribePatchBaselinesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribePatchBaselinesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchBaselinesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchBaselinesRequest(*ssm.DescribePatchBaselinesInput) (*request.Request, *ssm.DescribePatchBaselinesOutput) {
	panic("unimplemented")
}

// DescribePatchBaselinesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchBaselinesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchBaselinesWithContext(context.Context, *ssm.DescribePatchBaselinesInput, ...request.Option) (*ssm.DescribePatchBaselinesOutput, error) {
	panic("unimplemented")
}

// DescribePatchGroupState implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupState of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupState(*ssm.DescribePatchGroupStateInput) (*ssm.DescribePatchGroupStateOutput, error) {
	panic("unimplemented")
}

// DescribePatchGroupStateRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupStateRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupStateRequest(*ssm.DescribePatchGroupStateInput) (*request.Request, *ssm.DescribePatchGroupStateOutput) {
	panic("unimplemented")
}

// DescribePatchGroupStateWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupStateWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupStateWithContext(context.Context, *ssm.DescribePatchGroupStateInput, ...request.Option) (*ssm.DescribePatchGroupStateOutput, error) {
	panic("unimplemented")
}

// DescribePatchGroups implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroups of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroups(*ssm.DescribePatchGroupsInput) (*ssm.DescribePatchGroupsOutput, error) {
	panic("unimplemented")
}

// DescribePatchGroupsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupsPages(*ssm.DescribePatchGroupsInput, func(*ssm.DescribePatchGroupsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribePatchGroupsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupsPagesWithContext(context.Context, *ssm.DescribePatchGroupsInput, func(*ssm.DescribePatchGroupsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribePatchGroupsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupsRequest(*ssm.DescribePatchGroupsInput) (*request.Request, *ssm.DescribePatchGroupsOutput) {
	panic("unimplemented")
}

// DescribePatchGroupsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchGroupsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchGroupsWithContext(context.Context, *ssm.DescribePatchGroupsInput, ...request.Option) (*ssm.DescribePatchGroupsOutput, error) {
	panic("unimplemented")
}

// DescribePatchProperties implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchProperties of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchProperties(*ssm.DescribePatchPropertiesInput) (*ssm.DescribePatchPropertiesOutput, error) {
	panic("unimplemented")
}

// DescribePatchPropertiesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchPropertiesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchPropertiesPages(*ssm.DescribePatchPropertiesInput, func(*ssm.DescribePatchPropertiesOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribePatchPropertiesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchPropertiesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchPropertiesPagesWithContext(context.Context, *ssm.DescribePatchPropertiesInput, func(*ssm.DescribePatchPropertiesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribePatchPropertiesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchPropertiesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchPropertiesRequest(*ssm.DescribePatchPropertiesInput) (*request.Request, *ssm.DescribePatchPropertiesOutput) {
	panic("unimplemented")
}

// DescribePatchPropertiesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribePatchPropertiesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribePatchPropertiesWithContext(context.Context, *ssm.DescribePatchPropertiesInput, ...request.Option) (*ssm.DescribePatchPropertiesOutput, error) {
	panic("unimplemented")
}

// DescribeSessions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeSessions of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeSessions(*ssm.DescribeSessionsInput) (*ssm.DescribeSessionsOutput, error) {
	panic("unimplemented")
}

// DescribeSessionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeSessionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeSessionsPages(*ssm.DescribeSessionsInput, func(*ssm.DescribeSessionsOutput, bool) bool) error {
	panic("unimplemented")
}

// DescribeSessionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeSessionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeSessionsPagesWithContext(context.Context, *ssm.DescribeSessionsInput, func(*ssm.DescribeSessionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// DescribeSessionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeSessionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeSessionsRequest(*ssm.DescribeSessionsInput) (*request.Request, *ssm.DescribeSessionsOutput) {
	panic("unimplemented")
}

// DescribeSessionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DescribeSessionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DescribeSessionsWithContext(context.Context, *ssm.DescribeSessionsInput, ...request.Option) (*ssm.DescribeSessionsOutput, error) {
	panic("unimplemented")
}

// DisassociateOpsItemRelatedItem implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DisassociateOpsItemRelatedItem of mockedSSM.SSMAPI.
func (m *mockedSSM) DisassociateOpsItemRelatedItem(*ssm.DisassociateOpsItemRelatedItemInput) (*ssm.DisassociateOpsItemRelatedItemOutput, error) {
	panic("unimplemented")
}

// DisassociateOpsItemRelatedItemRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DisassociateOpsItemRelatedItemRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) DisassociateOpsItemRelatedItemRequest(*ssm.DisassociateOpsItemRelatedItemInput) (*request.Request, *ssm.DisassociateOpsItemRelatedItemOutput) {
	panic("unimplemented")
}

// DisassociateOpsItemRelatedItemWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).DisassociateOpsItemRelatedItemWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) DisassociateOpsItemRelatedItemWithContext(context.Context, *ssm.DisassociateOpsItemRelatedItemInput, ...request.Option) (*ssm.DisassociateOpsItemRelatedItemOutput, error) {
	panic("unimplemented")
}

// GetAutomationExecution implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetAutomationExecution of mockedSSM.SSMAPI.
func (m *mockedSSM) GetAutomationExecution(*ssm.GetAutomationExecutionInput) (*ssm.GetAutomationExecutionOutput, error) {
	panic("unimplemented")
}

// GetAutomationExecutionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetAutomationExecutionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetAutomationExecutionRequest(*ssm.GetAutomationExecutionInput) (*request.Request, *ssm.GetAutomationExecutionOutput) {
	panic("unimplemented")
}

// GetAutomationExecutionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetAutomationExecutionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetAutomationExecutionWithContext(context.Context, *ssm.GetAutomationExecutionInput, ...request.Option) (*ssm.GetAutomationExecutionOutput, error) {
	panic("unimplemented")
}

// GetCalendarState implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetCalendarState of mockedSSM.SSMAPI.
func (m *mockedSSM) GetCalendarState(*ssm.GetCalendarStateInput) (*ssm.GetCalendarStateOutput, error) {
	panic("unimplemented")
}

// GetCalendarStateRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetCalendarStateRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetCalendarStateRequest(*ssm.GetCalendarStateInput) (*request.Request, *ssm.GetCalendarStateOutput) {
	panic("unimplemented")
}

// GetCalendarStateWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetCalendarStateWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetCalendarStateWithContext(context.Context, *ssm.GetCalendarStateInput, ...request.Option) (*ssm.GetCalendarStateOutput, error) {
	panic("unimplemented")
}

// GetCommandInvocation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetCommandInvocation of mockedSSM.SSMAPI.
func (m *mockedSSM) GetCommandInvocation(*ssm.GetCommandInvocationInput) (*ssm.GetCommandInvocationOutput, error) {
	panic("unimplemented")
}

// GetCommandInvocationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetCommandInvocationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetCommandInvocationRequest(*ssm.GetCommandInvocationInput) (*request.Request, *ssm.GetCommandInvocationOutput) {
	panic("unimplemented")
}

// GetCommandInvocationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetCommandInvocationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetCommandInvocationWithContext(context.Context, *ssm.GetCommandInvocationInput, ...request.Option) (*ssm.GetCommandInvocationOutput, error) {
	panic("unimplemented")
}

// GetConnectionStatus implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetConnectionStatus of mockedSSM.SSMAPI.
func (m *mockedSSM) GetConnectionStatus(*ssm.GetConnectionStatusInput) (*ssm.GetConnectionStatusOutput, error) {
	panic("unimplemented")
}

// GetConnectionStatusRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetConnectionStatusRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetConnectionStatusRequest(*ssm.GetConnectionStatusInput) (*request.Request, *ssm.GetConnectionStatusOutput) {
	panic("unimplemented")
}

// GetConnectionStatusWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetConnectionStatusWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetConnectionStatusWithContext(context.Context, *ssm.GetConnectionStatusInput, ...request.Option) (*ssm.GetConnectionStatusOutput, error) {
	panic("unimplemented")
}

// GetDefaultPatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDefaultPatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDefaultPatchBaseline(*ssm.GetDefaultPatchBaselineInput) (*ssm.GetDefaultPatchBaselineOutput, error) {
	panic("unimplemented")
}

// GetDefaultPatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDefaultPatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDefaultPatchBaselineRequest(*ssm.GetDefaultPatchBaselineInput) (*request.Request, *ssm.GetDefaultPatchBaselineOutput) {
	panic("unimplemented")
}

// GetDefaultPatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDefaultPatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDefaultPatchBaselineWithContext(context.Context, *ssm.GetDefaultPatchBaselineInput, ...request.Option) (*ssm.GetDefaultPatchBaselineOutput, error) {
	panic("unimplemented")
}

// GetDeployablePatchSnapshotForInstance implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDeployablePatchSnapshotForInstance of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDeployablePatchSnapshotForInstance(*ssm.GetDeployablePatchSnapshotForInstanceInput) (*ssm.GetDeployablePatchSnapshotForInstanceOutput, error) {
	panic("unimplemented")
}

// GetDeployablePatchSnapshotForInstanceRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDeployablePatchSnapshotForInstanceRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDeployablePatchSnapshotForInstanceRequest(*ssm.GetDeployablePatchSnapshotForInstanceInput) (*request.Request, *ssm.GetDeployablePatchSnapshotForInstanceOutput) {
	panic("unimplemented")
}

// GetDeployablePatchSnapshotForInstanceWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDeployablePatchSnapshotForInstanceWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDeployablePatchSnapshotForInstanceWithContext(context.Context, *ssm.GetDeployablePatchSnapshotForInstanceInput, ...request.Option) (*ssm.GetDeployablePatchSnapshotForInstanceOutput, error) {
	panic("unimplemented")
}

// GetDocument implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDocument of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDocument(*ssm.GetDocumentInput) (*ssm.GetDocumentOutput, error) {
	panic("unimplemented")
}

// GetDocumentRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDocumentRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDocumentRequest(*ssm.GetDocumentInput) (*request.Request, *ssm.GetDocumentOutput) {
	panic("unimplemented")
}

// GetDocumentWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetDocumentWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetDocumentWithContext(context.Context, *ssm.GetDocumentInput, ...request.Option) (*ssm.GetDocumentOutput, error) {
	panic("unimplemented")
}

// GetInventory implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventory of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventory(*ssm.GetInventoryInput) (*ssm.GetInventoryOutput, error) {
	panic("unimplemented")
}

// GetInventoryPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventoryPages of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventoryPages(*ssm.GetInventoryInput, func(*ssm.GetInventoryOutput, bool) bool) error {
	panic("unimplemented")
}

// GetInventoryPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventoryPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventoryPagesWithContext(context.Context, *ssm.GetInventoryInput, func(*ssm.GetInventoryOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// GetInventoryRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventoryRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventoryRequest(*ssm.GetInventoryInput) (*request.Request, *ssm.GetInventoryOutput) {
	panic("unimplemented")
}

// GetInventorySchema implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventorySchema of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventorySchema(*ssm.GetInventorySchemaInput) (*ssm.GetInventorySchemaOutput, error) {
	panic("unimplemented")
}

// GetInventorySchemaPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventorySchemaPages of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventorySchemaPages(*ssm.GetInventorySchemaInput, func(*ssm.GetInventorySchemaOutput, bool) bool) error {
	panic("unimplemented")
}

// GetInventorySchemaPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventorySchemaPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventorySchemaPagesWithContext(context.Context, *ssm.GetInventorySchemaInput, func(*ssm.GetInventorySchemaOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// GetInventorySchemaRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventorySchemaRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventorySchemaRequest(*ssm.GetInventorySchemaInput) (*request.Request, *ssm.GetInventorySchemaOutput) {
	panic("unimplemented")
}

// GetInventorySchemaWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventorySchemaWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventorySchemaWithContext(context.Context, *ssm.GetInventorySchemaInput, ...request.Option) (*ssm.GetInventorySchemaOutput, error) {
	panic("unimplemented")
}

// GetInventoryWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetInventoryWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetInventoryWithContext(context.Context, *ssm.GetInventoryInput, ...request.Option) (*ssm.GetInventoryOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindow(*ssm.GetMaintenanceWindowInput) (*ssm.GetMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecution implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecution of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecution(*ssm.GetMaintenanceWindowExecutionInput) (*ssm.GetMaintenanceWindowExecutionOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionRequest(*ssm.GetMaintenanceWindowExecutionInput) (*request.Request, *ssm.GetMaintenanceWindowExecutionOutput) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionTask implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionTask of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionTask(*ssm.GetMaintenanceWindowExecutionTaskInput) (*ssm.GetMaintenanceWindowExecutionTaskOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionTaskInvocation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionTaskInvocation of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionTaskInvocation(*ssm.GetMaintenanceWindowExecutionTaskInvocationInput) (*ssm.GetMaintenanceWindowExecutionTaskInvocationOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionTaskInvocationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionTaskInvocationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionTaskInvocationRequest(*ssm.GetMaintenanceWindowExecutionTaskInvocationInput) (*request.Request, *ssm.GetMaintenanceWindowExecutionTaskInvocationOutput) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionTaskInvocationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionTaskInvocationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionTaskInvocationWithContext(context.Context, *ssm.GetMaintenanceWindowExecutionTaskInvocationInput, ...request.Option) (*ssm.GetMaintenanceWindowExecutionTaskInvocationOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionTaskRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionTaskRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionTaskRequest(*ssm.GetMaintenanceWindowExecutionTaskInput) (*request.Request, *ssm.GetMaintenanceWindowExecutionTaskOutput) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionTaskWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionTaskWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionTaskWithContext(context.Context, *ssm.GetMaintenanceWindowExecutionTaskInput, ...request.Option) (*ssm.GetMaintenanceWindowExecutionTaskOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowExecutionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowExecutionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowExecutionWithContext(context.Context, *ssm.GetMaintenanceWindowExecutionInput, ...request.Option) (*ssm.GetMaintenanceWindowExecutionOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowRequest(*ssm.GetMaintenanceWindowInput) (*request.Request, *ssm.GetMaintenanceWindowOutput) {
	panic("unimplemented")
}

// GetMaintenanceWindowTask implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowTask of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowTask(*ssm.GetMaintenanceWindowTaskInput) (*ssm.GetMaintenanceWindowTaskOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowTaskRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowTaskRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowTaskRequest(*ssm.GetMaintenanceWindowTaskInput) (*request.Request, *ssm.GetMaintenanceWindowTaskOutput) {
	panic("unimplemented")
}

// GetMaintenanceWindowTaskWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowTaskWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowTaskWithContext(context.Context, *ssm.GetMaintenanceWindowTaskInput, ...request.Option) (*ssm.GetMaintenanceWindowTaskOutput, error) {
	panic("unimplemented")
}

// GetMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetMaintenanceWindowWithContext(context.Context, *ssm.GetMaintenanceWindowInput, ...request.Option) (*ssm.GetMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// GetOpsItem implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsItem of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsItem(*ssm.GetOpsItemInput) (*ssm.GetOpsItemOutput, error) {
	panic("unimplemented")
}

// GetOpsItemRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsItemRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsItemRequest(*ssm.GetOpsItemInput) (*request.Request, *ssm.GetOpsItemOutput) {
	panic("unimplemented")
}

// GetOpsItemWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsItemWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsItemWithContext(context.Context, *ssm.GetOpsItemInput, ...request.Option) (*ssm.GetOpsItemOutput, error) {
	panic("unimplemented")
}

// GetOpsMetadata implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsMetadata of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsMetadata(*ssm.GetOpsMetadataInput) (*ssm.GetOpsMetadataOutput, error) {
	panic("unimplemented")
}

// GetOpsMetadataRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsMetadataRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsMetadataRequest(*ssm.GetOpsMetadataInput) (*request.Request, *ssm.GetOpsMetadataOutput) {
	panic("unimplemented")
}

// GetOpsMetadataWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsMetadataWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsMetadataWithContext(context.Context, *ssm.GetOpsMetadataInput, ...request.Option) (*ssm.GetOpsMetadataOutput, error) {
	panic("unimplemented")
}

// GetOpsSummary implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsSummary of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsSummary(*ssm.GetOpsSummaryInput) (*ssm.GetOpsSummaryOutput, error) {
	panic("unimplemented")
}

// GetOpsSummaryPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsSummaryPages of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsSummaryPages(*ssm.GetOpsSummaryInput, func(*ssm.GetOpsSummaryOutput, bool) bool) error {
	panic("unimplemented")
}

// GetOpsSummaryPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsSummaryPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsSummaryPagesWithContext(context.Context, *ssm.GetOpsSummaryInput, func(*ssm.GetOpsSummaryOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// GetOpsSummaryRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsSummaryRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsSummaryRequest(*ssm.GetOpsSummaryInput) (*request.Request, *ssm.GetOpsSummaryOutput) {
	panic("unimplemented")
}

// GetOpsSummaryWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetOpsSummaryWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetOpsSummaryWithContext(context.Context, *ssm.GetOpsSummaryInput, ...request.Option) (*ssm.GetOpsSummaryOutput, error) {
	panic("unimplemented")
}

// GetParameterHistory implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterHistory of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterHistory(*ssm.GetParameterHistoryInput) (*ssm.GetParameterHistoryOutput, error) {
	panic("unimplemented")
}

// GetParameterHistoryPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterHistoryPages of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterHistoryPages(*ssm.GetParameterHistoryInput, func(*ssm.GetParameterHistoryOutput, bool) bool) error {
	panic("unimplemented")
}

// GetParameterHistoryPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterHistoryPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterHistoryPagesWithContext(context.Context, *ssm.GetParameterHistoryInput, func(*ssm.GetParameterHistoryOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// GetParameterHistoryRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterHistoryRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterHistoryRequest(*ssm.GetParameterHistoryInput) (*request.Request, *ssm.GetParameterHistoryOutput) {
	panic("unimplemented")
}

// GetParameterHistoryWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterHistoryWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterHistoryWithContext(context.Context, *ssm.GetParameterHistoryInput, ...request.Option) (*ssm.GetParameterHistoryOutput, error) {
	panic("unimplemented")
}

// GetParameterRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterRequest(*ssm.GetParameterInput) (*request.Request, *ssm.GetParameterOutput) {
	panic("unimplemented")
}

// GetParameterWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameterWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameterWithContext(context.Context, *ssm.GetParameterInput, ...request.Option) (*ssm.GetParameterOutput, error) {
	panic("unimplemented")
}

// GetParameters implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParameters of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParameters(*ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	panic("unimplemented")
}

// GetParametersByPath implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersByPath of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersByPath(*ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	panic("unimplemented")
}

// GetParametersByPathPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersByPathPages of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersByPathPages(*ssm.GetParametersByPathInput, func(*ssm.GetParametersByPathOutput, bool) bool) error {
	panic("unimplemented")
}

// GetParametersByPathPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersByPathPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersByPathPagesWithContext(context.Context, *ssm.GetParametersByPathInput, func(*ssm.GetParametersByPathOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// GetParametersByPathRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersByPathRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersByPathRequest(*ssm.GetParametersByPathInput) (*request.Request, *ssm.GetParametersByPathOutput) {
	panic("unimplemented")
}

// GetParametersByPathWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersByPathWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersByPathWithContext(context.Context, *ssm.GetParametersByPathInput, ...request.Option) (*ssm.GetParametersByPathOutput, error) {
	panic("unimplemented")
}

// GetParametersRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersRequest(*ssm.GetParametersInput) (*request.Request, *ssm.GetParametersOutput) {
	panic("unimplemented")
}

// GetParametersWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetParametersWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetParametersWithContext(context.Context, *ssm.GetParametersInput, ...request.Option) (*ssm.GetParametersOutput, error) {
	panic("unimplemented")
}

// GetPatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetPatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) GetPatchBaseline(*ssm.GetPatchBaselineInput) (*ssm.GetPatchBaselineOutput, error) {
	panic("unimplemented")
}

// GetPatchBaselineForPatchGroup implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetPatchBaselineForPatchGroup of mockedSSM.SSMAPI.
func (m *mockedSSM) GetPatchBaselineForPatchGroup(*ssm.GetPatchBaselineForPatchGroupInput) (*ssm.GetPatchBaselineForPatchGroupOutput, error) {
	panic("unimplemented")
}

// GetPatchBaselineForPatchGroupRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetPatchBaselineForPatchGroupRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetPatchBaselineForPatchGroupRequest(*ssm.GetPatchBaselineForPatchGroupInput) (*request.Request, *ssm.GetPatchBaselineForPatchGroupOutput) {
	panic("unimplemented")
}

// GetPatchBaselineForPatchGroupWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetPatchBaselineForPatchGroupWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetPatchBaselineForPatchGroupWithContext(context.Context, *ssm.GetPatchBaselineForPatchGroupInput, ...request.Option) (*ssm.GetPatchBaselineForPatchGroupOutput, error) {
	panic("unimplemented")
}

// GetPatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetPatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetPatchBaselineRequest(*ssm.GetPatchBaselineInput) (*request.Request, *ssm.GetPatchBaselineOutput) {
	panic("unimplemented")
}

// GetPatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetPatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetPatchBaselineWithContext(context.Context, *ssm.GetPatchBaselineInput, ...request.Option) (*ssm.GetPatchBaselineOutput, error) {
	panic("unimplemented")
}

// GetResourcePolicies implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetResourcePolicies of mockedSSM.SSMAPI.
func (m *mockedSSM) GetResourcePolicies(*ssm.GetResourcePoliciesInput) (*ssm.GetResourcePoliciesOutput, error) {
	panic("unimplemented")
}

// GetResourcePoliciesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetResourcePoliciesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) GetResourcePoliciesPages(*ssm.GetResourcePoliciesInput, func(*ssm.GetResourcePoliciesOutput, bool) bool) error {
	panic("unimplemented")
}

// GetResourcePoliciesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetResourcePoliciesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetResourcePoliciesPagesWithContext(context.Context, *ssm.GetResourcePoliciesInput, func(*ssm.GetResourcePoliciesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// GetResourcePoliciesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetResourcePoliciesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetResourcePoliciesRequest(*ssm.GetResourcePoliciesInput) (*request.Request, *ssm.GetResourcePoliciesOutput) {
	panic("unimplemented")
}

// GetResourcePoliciesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetResourcePoliciesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetResourcePoliciesWithContext(context.Context, *ssm.GetResourcePoliciesInput, ...request.Option) (*ssm.GetResourcePoliciesOutput, error) {
	panic("unimplemented")
}

// GetServiceSetting implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetServiceSetting of mockedSSM.SSMAPI.
func (m *mockedSSM) GetServiceSetting(*ssm.GetServiceSettingInput) (*ssm.GetServiceSettingOutput, error) {
	panic("unimplemented")
}

// GetServiceSettingRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetServiceSettingRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) GetServiceSettingRequest(*ssm.GetServiceSettingInput) (*request.Request, *ssm.GetServiceSettingOutput) {
	panic("unimplemented")
}

// GetServiceSettingWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).GetServiceSettingWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) GetServiceSettingWithContext(context.Context, *ssm.GetServiceSettingInput, ...request.Option) (*ssm.GetServiceSettingOutput, error) {
	panic("unimplemented")
}

// LabelParameterVersion implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).LabelParameterVersion of mockedSSM.SSMAPI.
func (m *mockedSSM) LabelParameterVersion(*ssm.LabelParameterVersionInput) (*ssm.LabelParameterVersionOutput, error) {
	panic("unimplemented")
}

// LabelParameterVersionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).LabelParameterVersionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) LabelParameterVersionRequest(*ssm.LabelParameterVersionInput) (*request.Request, *ssm.LabelParameterVersionOutput) {
	panic("unimplemented")
}

// LabelParameterVersionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).LabelParameterVersionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) LabelParameterVersionWithContext(context.Context, *ssm.LabelParameterVersionInput, ...request.Option) (*ssm.LabelParameterVersionOutput, error) {
	panic("unimplemented")
}

// ListAssociationVersions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationVersions of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationVersions(*ssm.ListAssociationVersionsInput) (*ssm.ListAssociationVersionsOutput, error) {
	panic("unimplemented")
}

// ListAssociationVersionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationVersionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationVersionsPages(*ssm.ListAssociationVersionsInput, func(*ssm.ListAssociationVersionsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListAssociationVersionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationVersionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationVersionsPagesWithContext(context.Context, *ssm.ListAssociationVersionsInput, func(*ssm.ListAssociationVersionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListAssociationVersionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationVersionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationVersionsRequest(*ssm.ListAssociationVersionsInput) (*request.Request, *ssm.ListAssociationVersionsOutput) {
	panic("unimplemented")
}

// ListAssociationVersionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationVersionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationVersionsWithContext(context.Context, *ssm.ListAssociationVersionsInput, ...request.Option) (*ssm.ListAssociationVersionsOutput, error) {
	panic("unimplemented")
}

// ListAssociations implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociations of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociations(*ssm.ListAssociationsInput) (*ssm.ListAssociationsOutput, error) {
	panic("unimplemented")
}

// ListAssociationsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationsPages(*ssm.ListAssociationsInput, func(*ssm.ListAssociationsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListAssociationsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationsPagesWithContext(context.Context, *ssm.ListAssociationsInput, func(*ssm.ListAssociationsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListAssociationsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationsRequest(*ssm.ListAssociationsInput) (*request.Request, *ssm.ListAssociationsOutput) {
	panic("unimplemented")
}

// ListAssociationsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListAssociationsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListAssociationsWithContext(context.Context, *ssm.ListAssociationsInput, ...request.Option) (*ssm.ListAssociationsOutput, error) {
	panic("unimplemented")
}

// ListCommandInvocations implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandInvocations of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandInvocations(*ssm.ListCommandInvocationsInput) (*ssm.ListCommandInvocationsOutput, error) {
	panic("unimplemented")
}

// ListCommandInvocationsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandInvocationsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandInvocationsPages(*ssm.ListCommandInvocationsInput, func(*ssm.ListCommandInvocationsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListCommandInvocationsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandInvocationsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandInvocationsPagesWithContext(context.Context, *ssm.ListCommandInvocationsInput, func(*ssm.ListCommandInvocationsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListCommandInvocationsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandInvocationsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandInvocationsRequest(*ssm.ListCommandInvocationsInput) (*request.Request, *ssm.ListCommandInvocationsOutput) {
	panic("unimplemented")
}

// ListCommandInvocationsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandInvocationsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandInvocationsWithContext(context.Context, *ssm.ListCommandInvocationsInput, ...request.Option) (*ssm.ListCommandInvocationsOutput, error) {
	panic("unimplemented")
}

// ListCommands implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommands of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommands(*ssm.ListCommandsInput) (*ssm.ListCommandsOutput, error) {
	panic("unimplemented")
}

// ListCommandsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandsPages(*ssm.ListCommandsInput, func(*ssm.ListCommandsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListCommandsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandsPagesWithContext(context.Context, *ssm.ListCommandsInput, func(*ssm.ListCommandsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListCommandsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandsRequest(*ssm.ListCommandsInput) (*request.Request, *ssm.ListCommandsOutput) {
	panic("unimplemented")
}

// ListCommandsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListCommandsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListCommandsWithContext(context.Context, *ssm.ListCommandsInput, ...request.Option) (*ssm.ListCommandsOutput, error) {
	panic("unimplemented")
}

// ListComplianceItems implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceItems of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceItems(*ssm.ListComplianceItemsInput) (*ssm.ListComplianceItemsOutput, error) {
	panic("unimplemented")
}

// ListComplianceItemsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceItemsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceItemsPages(*ssm.ListComplianceItemsInput, func(*ssm.ListComplianceItemsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListComplianceItemsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceItemsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceItemsPagesWithContext(context.Context, *ssm.ListComplianceItemsInput, func(*ssm.ListComplianceItemsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListComplianceItemsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceItemsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceItemsRequest(*ssm.ListComplianceItemsInput) (*request.Request, *ssm.ListComplianceItemsOutput) {
	panic("unimplemented")
}

// ListComplianceItemsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceItemsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceItemsWithContext(context.Context, *ssm.ListComplianceItemsInput, ...request.Option) (*ssm.ListComplianceItemsOutput, error) {
	panic("unimplemented")
}

// ListComplianceSummaries implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceSummaries of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceSummaries(*ssm.ListComplianceSummariesInput) (*ssm.ListComplianceSummariesOutput, error) {
	panic("unimplemented")
}

// ListComplianceSummariesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceSummariesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceSummariesPages(*ssm.ListComplianceSummariesInput, func(*ssm.ListComplianceSummariesOutput, bool) bool) error {
	panic("unimplemented")
}

// ListComplianceSummariesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceSummariesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceSummariesPagesWithContext(context.Context, *ssm.ListComplianceSummariesInput, func(*ssm.ListComplianceSummariesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListComplianceSummariesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceSummariesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceSummariesRequest(*ssm.ListComplianceSummariesInput) (*request.Request, *ssm.ListComplianceSummariesOutput) {
	panic("unimplemented")
}

// ListComplianceSummariesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListComplianceSummariesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListComplianceSummariesWithContext(context.Context, *ssm.ListComplianceSummariesInput, ...request.Option) (*ssm.ListComplianceSummariesOutput, error) {
	panic("unimplemented")
}

// ListDocumentMetadataHistory implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentMetadataHistory of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentMetadataHistory(*ssm.ListDocumentMetadataHistoryInput) (*ssm.ListDocumentMetadataHistoryOutput, error) {
	panic("unimplemented")
}

// ListDocumentMetadataHistoryRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentMetadataHistoryRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentMetadataHistoryRequest(*ssm.ListDocumentMetadataHistoryInput) (*request.Request, *ssm.ListDocumentMetadataHistoryOutput) {
	panic("unimplemented")
}

// ListDocumentMetadataHistoryWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentMetadataHistoryWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentMetadataHistoryWithContext(context.Context, *ssm.ListDocumentMetadataHistoryInput, ...request.Option) (*ssm.ListDocumentMetadataHistoryOutput, error) {
	panic("unimplemented")
}

// ListDocumentVersions implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentVersions of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentVersions(*ssm.ListDocumentVersionsInput) (*ssm.ListDocumentVersionsOutput, error) {
	panic("unimplemented")
}

// ListDocumentVersionsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentVersionsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentVersionsPages(*ssm.ListDocumentVersionsInput, func(*ssm.ListDocumentVersionsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListDocumentVersionsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentVersionsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentVersionsPagesWithContext(context.Context, *ssm.ListDocumentVersionsInput, func(*ssm.ListDocumentVersionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListDocumentVersionsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentVersionsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentVersionsRequest(*ssm.ListDocumentVersionsInput) (*request.Request, *ssm.ListDocumentVersionsOutput) {
	panic("unimplemented")
}

// ListDocumentVersionsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentVersionsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentVersionsWithContext(context.Context, *ssm.ListDocumentVersionsInput, ...request.Option) (*ssm.ListDocumentVersionsOutput, error) {
	panic("unimplemented")
}

// ListDocuments implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocuments of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocuments(*ssm.ListDocumentsInput) (*ssm.ListDocumentsOutput, error) {
	panic("unimplemented")
}

// ListDocumentsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentsPages(*ssm.ListDocumentsInput, func(*ssm.ListDocumentsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListDocumentsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentsPagesWithContext(context.Context, *ssm.ListDocumentsInput, func(*ssm.ListDocumentsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListDocumentsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentsRequest(*ssm.ListDocumentsInput) (*request.Request, *ssm.ListDocumentsOutput) {
	panic("unimplemented")
}

// ListDocumentsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListDocumentsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListDocumentsWithContext(context.Context, *ssm.ListDocumentsInput, ...request.Option) (*ssm.ListDocumentsOutput, error) {
	panic("unimplemented")
}

// ListInventoryEntries implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListInventoryEntries of mockedSSM.SSMAPI.
func (m *mockedSSM) ListInventoryEntries(*ssm.ListInventoryEntriesInput) (*ssm.ListInventoryEntriesOutput, error) {
	panic("unimplemented")
}

// ListInventoryEntriesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListInventoryEntriesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListInventoryEntriesRequest(*ssm.ListInventoryEntriesInput) (*request.Request, *ssm.ListInventoryEntriesOutput) {
	panic("unimplemented")
}

// ListInventoryEntriesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListInventoryEntriesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListInventoryEntriesWithContext(context.Context, *ssm.ListInventoryEntriesInput, ...request.Option) (*ssm.ListInventoryEntriesOutput, error) {
	panic("unimplemented")
}

// ListOpsItemEvents implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemEvents of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemEvents(*ssm.ListOpsItemEventsInput) (*ssm.ListOpsItemEventsOutput, error) {
	panic("unimplemented")
}

// ListOpsItemEventsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemEventsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemEventsPages(*ssm.ListOpsItemEventsInput, func(*ssm.ListOpsItemEventsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListOpsItemEventsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemEventsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemEventsPagesWithContext(context.Context, *ssm.ListOpsItemEventsInput, func(*ssm.ListOpsItemEventsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListOpsItemEventsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemEventsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemEventsRequest(*ssm.ListOpsItemEventsInput) (*request.Request, *ssm.ListOpsItemEventsOutput) {
	panic("unimplemented")
}

// ListOpsItemEventsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemEventsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemEventsWithContext(context.Context, *ssm.ListOpsItemEventsInput, ...request.Option) (*ssm.ListOpsItemEventsOutput, error) {
	panic("unimplemented")
}

// ListOpsItemRelatedItems implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemRelatedItems of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemRelatedItems(*ssm.ListOpsItemRelatedItemsInput) (*ssm.ListOpsItemRelatedItemsOutput, error) {
	panic("unimplemented")
}

// ListOpsItemRelatedItemsPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemRelatedItemsPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemRelatedItemsPages(*ssm.ListOpsItemRelatedItemsInput, func(*ssm.ListOpsItemRelatedItemsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListOpsItemRelatedItemsPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemRelatedItemsPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemRelatedItemsPagesWithContext(context.Context, *ssm.ListOpsItemRelatedItemsInput, func(*ssm.ListOpsItemRelatedItemsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListOpsItemRelatedItemsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemRelatedItemsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemRelatedItemsRequest(*ssm.ListOpsItemRelatedItemsInput) (*request.Request, *ssm.ListOpsItemRelatedItemsOutput) {
	panic("unimplemented")
}

// ListOpsItemRelatedItemsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsItemRelatedItemsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsItemRelatedItemsWithContext(context.Context, *ssm.ListOpsItemRelatedItemsInput, ...request.Option) (*ssm.ListOpsItemRelatedItemsOutput, error) {
	panic("unimplemented")
}

// ListOpsMetadata implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsMetadata of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsMetadata(*ssm.ListOpsMetadataInput) (*ssm.ListOpsMetadataOutput, error) {
	panic("unimplemented")
}

// ListOpsMetadataPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsMetadataPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsMetadataPages(*ssm.ListOpsMetadataInput, func(*ssm.ListOpsMetadataOutput, bool) bool) error {
	panic("unimplemented")
}

// ListOpsMetadataPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsMetadataPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsMetadataPagesWithContext(context.Context, *ssm.ListOpsMetadataInput, func(*ssm.ListOpsMetadataOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListOpsMetadataRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsMetadataRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsMetadataRequest(*ssm.ListOpsMetadataInput) (*request.Request, *ssm.ListOpsMetadataOutput) {
	panic("unimplemented")
}

// ListOpsMetadataWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListOpsMetadataWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListOpsMetadataWithContext(context.Context, *ssm.ListOpsMetadataInput, ...request.Option) (*ssm.ListOpsMetadataOutput, error) {
	panic("unimplemented")
}

// ListResourceComplianceSummaries implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceComplianceSummaries of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceComplianceSummaries(*ssm.ListResourceComplianceSummariesInput) (*ssm.ListResourceComplianceSummariesOutput, error) {
	panic("unimplemented")
}

// ListResourceComplianceSummariesPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceComplianceSummariesPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceComplianceSummariesPages(*ssm.ListResourceComplianceSummariesInput, func(*ssm.ListResourceComplianceSummariesOutput, bool) bool) error {
	panic("unimplemented")
}

// ListResourceComplianceSummariesPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceComplianceSummariesPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceComplianceSummariesPagesWithContext(context.Context, *ssm.ListResourceComplianceSummariesInput, func(*ssm.ListResourceComplianceSummariesOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListResourceComplianceSummariesRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceComplianceSummariesRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceComplianceSummariesRequest(*ssm.ListResourceComplianceSummariesInput) (*request.Request, *ssm.ListResourceComplianceSummariesOutput) {
	panic("unimplemented")
}

// ListResourceComplianceSummariesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceComplianceSummariesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceComplianceSummariesWithContext(context.Context, *ssm.ListResourceComplianceSummariesInput, ...request.Option) (*ssm.ListResourceComplianceSummariesOutput, error) {
	panic("unimplemented")
}

// ListResourceDataSync implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceDataSync of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceDataSync(*ssm.ListResourceDataSyncInput) (*ssm.ListResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// ListResourceDataSyncPages implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceDataSyncPages of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceDataSyncPages(*ssm.ListResourceDataSyncInput, func(*ssm.ListResourceDataSyncOutput, bool) bool) error {
	panic("unimplemented")
}

// ListResourceDataSyncPagesWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceDataSyncPagesWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceDataSyncPagesWithContext(context.Context, *ssm.ListResourceDataSyncInput, func(*ssm.ListResourceDataSyncOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListResourceDataSyncRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceDataSyncRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceDataSyncRequest(*ssm.ListResourceDataSyncInput) (*request.Request, *ssm.ListResourceDataSyncOutput) {
	panic("unimplemented")
}

// ListResourceDataSyncWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListResourceDataSyncWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListResourceDataSyncWithContext(context.Context, *ssm.ListResourceDataSyncInput, ...request.Option) (*ssm.ListResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// ListTagsForResource implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListTagsForResource of mockedSSM.SSMAPI.
func (m *mockedSSM) ListTagsForResource(*ssm.ListTagsForResourceInput) (*ssm.ListTagsForResourceOutput, error) {
	panic("unimplemented")
}

// ListTagsForResourceRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListTagsForResourceRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ListTagsForResourceRequest(*ssm.ListTagsForResourceInput) (*request.Request, *ssm.ListTagsForResourceOutput) {
	panic("unimplemented")
}

// ListTagsForResourceWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ListTagsForResourceWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ListTagsForResourceWithContext(context.Context, *ssm.ListTagsForResourceInput, ...request.Option) (*ssm.ListTagsForResourceOutput, error) {
	panic("unimplemented")
}

// ModifyDocumentPermission implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ModifyDocumentPermission of mockedSSM.SSMAPI.
func (m *mockedSSM) ModifyDocumentPermission(*ssm.ModifyDocumentPermissionInput) (*ssm.ModifyDocumentPermissionOutput, error) {
	panic("unimplemented")
}

// ModifyDocumentPermissionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ModifyDocumentPermissionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ModifyDocumentPermissionRequest(*ssm.ModifyDocumentPermissionInput) (*request.Request, *ssm.ModifyDocumentPermissionOutput) {
	panic("unimplemented")
}

// ModifyDocumentPermissionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ModifyDocumentPermissionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ModifyDocumentPermissionWithContext(context.Context, *ssm.ModifyDocumentPermissionInput, ...request.Option) (*ssm.ModifyDocumentPermissionOutput, error) {
	panic("unimplemented")
}

// PutComplianceItems implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutComplianceItems of mockedSSM.SSMAPI.
func (m *mockedSSM) PutComplianceItems(*ssm.PutComplianceItemsInput) (*ssm.PutComplianceItemsOutput, error) {
	panic("unimplemented")
}

// PutComplianceItemsRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutComplianceItemsRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) PutComplianceItemsRequest(*ssm.PutComplianceItemsInput) (*request.Request, *ssm.PutComplianceItemsOutput) {
	panic("unimplemented")
}

// PutComplianceItemsWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutComplianceItemsWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) PutComplianceItemsWithContext(context.Context, *ssm.PutComplianceItemsInput, ...request.Option) (*ssm.PutComplianceItemsOutput, error) {
	panic("unimplemented")
}

// PutInventory implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutInventory of mockedSSM.SSMAPI.
func (m *mockedSSM) PutInventory(*ssm.PutInventoryInput) (*ssm.PutInventoryOutput, error) {
	panic("unimplemented")
}

// PutInventoryRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutInventoryRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) PutInventoryRequest(*ssm.PutInventoryInput) (*request.Request, *ssm.PutInventoryOutput) {
	panic("unimplemented")
}

// PutInventoryWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutInventoryWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) PutInventoryWithContext(context.Context, *ssm.PutInventoryInput, ...request.Option) (*ssm.PutInventoryOutput, error) {
	panic("unimplemented")
}

// PutParameter implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutParameter of mockedSSM.SSMAPI.
func (m *mockedSSM) PutParameter(*ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	panic("unimplemented")
}

// PutParameterRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutParameterRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) PutParameterRequest(*ssm.PutParameterInput) (*request.Request, *ssm.PutParameterOutput) {
	panic("unimplemented")
}

// PutParameterWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).PutParameterWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) PutParameterWithContext(context.Context, *ssm.PutParameterInput, ...request.Option) (*ssm.PutParameterOutput, error) {
	panic("unimplemented")
}

// PutResourcePolicy implements ssmiface.SSMAPI.
func (m *mockedSSM) PutResourcePolicy(*ssm.PutResourcePolicyInput) (*ssm.PutResourcePolicyOutput, error) {
	panic("unimplemented")
}

// PutResourcePolicyRequest implements ssmiface.SSMAPI.
func (m *mockedSSM) PutResourcePolicyRequest(*ssm.PutResourcePolicyInput) (*request.Request, *ssm.PutResourcePolicyOutput) {
	panic("unimplemented")
}

// PutResourcePolicyWithContext implements ssmiface.SSMAPI.
func (m *mockedSSM) PutResourcePolicyWithContext(context.Context, *ssm.PutResourcePolicyInput, ...request.Option) (*ssm.PutResourcePolicyOutput, error) {
	panic("unimplemented")
}

// RegisterDefaultPatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterDefaultPatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterDefaultPatchBaseline(*ssm.RegisterDefaultPatchBaselineInput) (*ssm.RegisterDefaultPatchBaselineOutput, error) {
	panic("unimplemented")
}

// RegisterDefaultPatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterDefaultPatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterDefaultPatchBaselineRequest(*ssm.RegisterDefaultPatchBaselineInput) (*request.Request, *ssm.RegisterDefaultPatchBaselineOutput) {
	panic("unimplemented")
}

// RegisterDefaultPatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterDefaultPatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterDefaultPatchBaselineWithContext(context.Context, *ssm.RegisterDefaultPatchBaselineInput, ...request.Option) (*ssm.RegisterDefaultPatchBaselineOutput, error) {
	panic("unimplemented")
}

// RegisterPatchBaselineForPatchGroup implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterPatchBaselineForPatchGroup of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterPatchBaselineForPatchGroup(*ssm.RegisterPatchBaselineForPatchGroupInput) (*ssm.RegisterPatchBaselineForPatchGroupOutput, error) {
	panic("unimplemented")
}

// RegisterPatchBaselineForPatchGroupRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterPatchBaselineForPatchGroupRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterPatchBaselineForPatchGroupRequest(*ssm.RegisterPatchBaselineForPatchGroupInput) (*request.Request, *ssm.RegisterPatchBaselineForPatchGroupOutput) {
	panic("unimplemented")
}

// RegisterPatchBaselineForPatchGroupWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterPatchBaselineForPatchGroupWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterPatchBaselineForPatchGroupWithContext(context.Context, *ssm.RegisterPatchBaselineForPatchGroupInput, ...request.Option) (*ssm.RegisterPatchBaselineForPatchGroupOutput, error) {
	panic("unimplemented")
}

// RegisterTargetWithMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterTargetWithMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterTargetWithMaintenanceWindow(*ssm.RegisterTargetWithMaintenanceWindowInput) (*ssm.RegisterTargetWithMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// RegisterTargetWithMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterTargetWithMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterTargetWithMaintenanceWindowRequest(*ssm.RegisterTargetWithMaintenanceWindowInput) (*request.Request, *ssm.RegisterTargetWithMaintenanceWindowOutput) {
	panic("unimplemented")
}

// RegisterTargetWithMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterTargetWithMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterTargetWithMaintenanceWindowWithContext(context.Context, *ssm.RegisterTargetWithMaintenanceWindowInput, ...request.Option) (*ssm.RegisterTargetWithMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// RegisterTaskWithMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterTaskWithMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterTaskWithMaintenanceWindow(*ssm.RegisterTaskWithMaintenanceWindowInput) (*ssm.RegisterTaskWithMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// RegisterTaskWithMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterTaskWithMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterTaskWithMaintenanceWindowRequest(*ssm.RegisterTaskWithMaintenanceWindowInput) (*request.Request, *ssm.RegisterTaskWithMaintenanceWindowOutput) {
	panic("unimplemented")
}

// RegisterTaskWithMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RegisterTaskWithMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) RegisterTaskWithMaintenanceWindowWithContext(context.Context, *ssm.RegisterTaskWithMaintenanceWindowInput, ...request.Option) (*ssm.RegisterTaskWithMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// RemoveTagsFromResource implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RemoveTagsFromResource of mockedSSM.SSMAPI.
func (m *mockedSSM) RemoveTagsFromResource(*ssm.RemoveTagsFromResourceInput) (*ssm.RemoveTagsFromResourceOutput, error) {
	panic("unimplemented")
}

// RemoveTagsFromResourceRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RemoveTagsFromResourceRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) RemoveTagsFromResourceRequest(*ssm.RemoveTagsFromResourceInput) (*request.Request, *ssm.RemoveTagsFromResourceOutput) {
	panic("unimplemented")
}

// RemoveTagsFromResourceWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).RemoveTagsFromResourceWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) RemoveTagsFromResourceWithContext(context.Context, *ssm.RemoveTagsFromResourceInput, ...request.Option) (*ssm.RemoveTagsFromResourceOutput, error) {
	panic("unimplemented")
}

// ResetServiceSetting implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ResetServiceSetting of mockedSSM.SSMAPI.
func (m *mockedSSM) ResetServiceSetting(*ssm.ResetServiceSettingInput) (*ssm.ResetServiceSettingOutput, error) {
	panic("unimplemented")
}

// ResetServiceSettingRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ResetServiceSettingRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ResetServiceSettingRequest(*ssm.ResetServiceSettingInput) (*request.Request, *ssm.ResetServiceSettingOutput) {
	panic("unimplemented")
}

// ResetServiceSettingWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ResetServiceSettingWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ResetServiceSettingWithContext(context.Context, *ssm.ResetServiceSettingInput, ...request.Option) (*ssm.ResetServiceSettingOutput, error) {
	panic("unimplemented")
}

// ResumeSession implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ResumeSession of mockedSSM.SSMAPI.
func (m *mockedSSM) ResumeSession(*ssm.ResumeSessionInput) (*ssm.ResumeSessionOutput, error) {
	panic("unimplemented")
}

// ResumeSessionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ResumeSessionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) ResumeSessionRequest(*ssm.ResumeSessionInput) (*request.Request, *ssm.ResumeSessionOutput) {
	panic("unimplemented")
}

// ResumeSessionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).ResumeSessionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) ResumeSessionWithContext(context.Context, *ssm.ResumeSessionInput, ...request.Option) (*ssm.ResumeSessionOutput, error) {
	panic("unimplemented")
}

// SendAutomationSignal implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).SendAutomationSignal of mockedSSM.SSMAPI.
func (m *mockedSSM) SendAutomationSignal(*ssm.SendAutomationSignalInput) (*ssm.SendAutomationSignalOutput, error) {
	panic("unimplemented")
}

// SendAutomationSignalRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).SendAutomationSignalRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) SendAutomationSignalRequest(*ssm.SendAutomationSignalInput) (*request.Request, *ssm.SendAutomationSignalOutput) {
	panic("unimplemented")
}

// SendAutomationSignalWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).SendAutomationSignalWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) SendAutomationSignalWithContext(context.Context, *ssm.SendAutomationSignalInput, ...request.Option) (*ssm.SendAutomationSignalOutput, error) {
	panic("unimplemented")
}

// SendCommand implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).SendCommand of mockedSSM.SSMAPI.
func (m *mockedSSM) SendCommand(*ssm.SendCommandInput) (*ssm.SendCommandOutput, error) {
	panic("unimplemented")
}

// SendCommandRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).SendCommandRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) SendCommandRequest(*ssm.SendCommandInput) (*request.Request, *ssm.SendCommandOutput) {
	panic("unimplemented")
}

// SendCommandWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).SendCommandWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) SendCommandWithContext(context.Context, *ssm.SendCommandInput, ...request.Option) (*ssm.SendCommandOutput, error) {
	panic("unimplemented")
}

// StartAssociationsOnce implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartAssociationsOnce of mockedSSM.SSMAPI.
func (m *mockedSSM) StartAssociationsOnce(*ssm.StartAssociationsOnceInput) (*ssm.StartAssociationsOnceOutput, error) {
	panic("unimplemented")
}

// StartAssociationsOnceRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartAssociationsOnceRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) StartAssociationsOnceRequest(*ssm.StartAssociationsOnceInput) (*request.Request, *ssm.StartAssociationsOnceOutput) {
	panic("unimplemented")
}

// StartAssociationsOnceWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartAssociationsOnceWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) StartAssociationsOnceWithContext(context.Context, *ssm.StartAssociationsOnceInput, ...request.Option) (*ssm.StartAssociationsOnceOutput, error) {
	panic("unimplemented")
}

// StartAutomationExecution implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartAutomationExecution of mockedSSM.SSMAPI.
func (m *mockedSSM) StartAutomationExecution(*ssm.StartAutomationExecutionInput) (*ssm.StartAutomationExecutionOutput, error) {
	panic("unimplemented")
}

// StartAutomationExecutionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartAutomationExecutionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) StartAutomationExecutionRequest(*ssm.StartAutomationExecutionInput) (*request.Request, *ssm.StartAutomationExecutionOutput) {
	panic("unimplemented")
}

// StartAutomationExecutionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartAutomationExecutionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) StartAutomationExecutionWithContext(context.Context, *ssm.StartAutomationExecutionInput, ...request.Option) (*ssm.StartAutomationExecutionOutput, error) {
	panic("unimplemented")
}

// StartChangeRequestExecution implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartChangeRequestExecution of mockedSSM.SSMAPI.
func (m *mockedSSM) StartChangeRequestExecution(*ssm.StartChangeRequestExecutionInput) (*ssm.StartChangeRequestExecutionOutput, error) {
	panic("unimplemented")
}

// StartChangeRequestExecutionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartChangeRequestExecutionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) StartChangeRequestExecutionRequest(*ssm.StartChangeRequestExecutionInput) (*request.Request, *ssm.StartChangeRequestExecutionOutput) {
	panic("unimplemented")
}

// StartChangeRequestExecutionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartChangeRequestExecutionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) StartChangeRequestExecutionWithContext(context.Context, *ssm.StartChangeRequestExecutionInput, ...request.Option) (*ssm.StartChangeRequestExecutionOutput, error) {
	panic("unimplemented")
}

// StartSession implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartSession of mockedSSM.SSMAPI.
func (m *mockedSSM) StartSession(*ssm.StartSessionInput) (*ssm.StartSessionOutput, error) {
	panic("unimplemented")
}

// StartSessionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartSessionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) StartSessionRequest(*ssm.StartSessionInput) (*request.Request, *ssm.StartSessionOutput) {
	panic("unimplemented")
}

// StartSessionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StartSessionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) StartSessionWithContext(context.Context, *ssm.StartSessionInput, ...request.Option) (*ssm.StartSessionOutput, error) {
	panic("unimplemented")
}

// StopAutomationExecution implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StopAutomationExecution of mockedSSM.SSMAPI.
func (m *mockedSSM) StopAutomationExecution(*ssm.StopAutomationExecutionInput) (*ssm.StopAutomationExecutionOutput, error) {
	panic("unimplemented")
}

// StopAutomationExecutionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StopAutomationExecutionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) StopAutomationExecutionRequest(*ssm.StopAutomationExecutionInput) (*request.Request, *ssm.StopAutomationExecutionOutput) {
	panic("unimplemented")
}

// StopAutomationExecutionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).StopAutomationExecutionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) StopAutomationExecutionWithContext(context.Context, *ssm.StopAutomationExecutionInput, ...request.Option) (*ssm.StopAutomationExecutionOutput, error) {
	panic("unimplemented")
}

// TerminateSession implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).TerminateSession of mockedSSM.SSMAPI.
func (m *mockedSSM) TerminateSession(*ssm.TerminateSessionInput) (*ssm.TerminateSessionOutput, error) {
	panic("unimplemented")
}

// TerminateSessionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).TerminateSessionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) TerminateSessionRequest(*ssm.TerminateSessionInput) (*request.Request, *ssm.TerminateSessionOutput) {
	panic("unimplemented")
}

// TerminateSessionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).TerminateSessionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) TerminateSessionWithContext(context.Context, *ssm.TerminateSessionInput, ...request.Option) (*ssm.TerminateSessionOutput, error) {
	panic("unimplemented")
}

// UnlabelParameterVersion implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UnlabelParameterVersion of mockedSSM.SSMAPI.
func (m *mockedSSM) UnlabelParameterVersion(*ssm.UnlabelParameterVersionInput) (*ssm.UnlabelParameterVersionOutput, error) {
	panic("unimplemented")
}

// UnlabelParameterVersionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UnlabelParameterVersionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UnlabelParameterVersionRequest(*ssm.UnlabelParameterVersionInput) (*request.Request, *ssm.UnlabelParameterVersionOutput) {
	panic("unimplemented")
}

// UnlabelParameterVersionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UnlabelParameterVersionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UnlabelParameterVersionWithContext(context.Context, *ssm.UnlabelParameterVersionInput, ...request.Option) (*ssm.UnlabelParameterVersionOutput, error) {
	panic("unimplemented")
}

// UpdateAssociation implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateAssociation of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateAssociation(*ssm.UpdateAssociationInput) (*ssm.UpdateAssociationOutput, error) {
	panic("unimplemented")
}

// UpdateAssociationRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateAssociationRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateAssociationRequest(*ssm.UpdateAssociationInput) (*request.Request, *ssm.UpdateAssociationOutput) {
	panic("unimplemented")
}

// UpdateAssociationStatus implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateAssociationStatus of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateAssociationStatus(*ssm.UpdateAssociationStatusInput) (*ssm.UpdateAssociationStatusOutput, error) {
	panic("unimplemented")
}

// UpdateAssociationStatusRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateAssociationStatusRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateAssociationStatusRequest(*ssm.UpdateAssociationStatusInput) (*request.Request, *ssm.UpdateAssociationStatusOutput) {
	panic("unimplemented")
}

// UpdateAssociationStatusWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateAssociationStatusWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateAssociationStatusWithContext(context.Context, *ssm.UpdateAssociationStatusInput, ...request.Option) (*ssm.UpdateAssociationStatusOutput, error) {
	panic("unimplemented")
}

// UpdateAssociationWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateAssociationWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateAssociationWithContext(context.Context, *ssm.UpdateAssociationInput, ...request.Option) (*ssm.UpdateAssociationOutput, error) {
	panic("unimplemented")
}

// UpdateDocument implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocument of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocument(*ssm.UpdateDocumentInput) (*ssm.UpdateDocumentOutput, error) {
	panic("unimplemented")
}

// UpdateDocumentDefaultVersion implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentDefaultVersion of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentDefaultVersion(*ssm.UpdateDocumentDefaultVersionInput) (*ssm.UpdateDocumentDefaultVersionOutput, error) {
	panic("unimplemented")
}

// UpdateDocumentDefaultVersionRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentDefaultVersionRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentDefaultVersionRequest(*ssm.UpdateDocumentDefaultVersionInput) (*request.Request, *ssm.UpdateDocumentDefaultVersionOutput) {
	panic("unimplemented")
}

// UpdateDocumentDefaultVersionWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentDefaultVersionWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentDefaultVersionWithContext(context.Context, *ssm.UpdateDocumentDefaultVersionInput, ...request.Option) (*ssm.UpdateDocumentDefaultVersionOutput, error) {
	panic("unimplemented")
}

// UpdateDocumentMetadata implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentMetadata of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentMetadata(*ssm.UpdateDocumentMetadataInput) (*ssm.UpdateDocumentMetadataOutput, error) {
	panic("unimplemented")
}

// UpdateDocumentMetadataRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentMetadataRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentMetadataRequest(*ssm.UpdateDocumentMetadataInput) (*request.Request, *ssm.UpdateDocumentMetadataOutput) {
	panic("unimplemented")
}

// UpdateDocumentMetadataWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentMetadataWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentMetadataWithContext(context.Context, *ssm.UpdateDocumentMetadataInput, ...request.Option) (*ssm.UpdateDocumentMetadataOutput, error) {
	panic("unimplemented")
}

// UpdateDocumentRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentRequest(*ssm.UpdateDocumentInput) (*request.Request, *ssm.UpdateDocumentOutput) {
	panic("unimplemented")
}

// UpdateDocumentWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateDocumentWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateDocumentWithContext(context.Context, *ssm.UpdateDocumentInput, ...request.Option) (*ssm.UpdateDocumentOutput, error) {
	panic("unimplemented")
}

// UpdateMaintenanceWindow implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindow of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindow(*ssm.UpdateMaintenanceWindowInput) (*ssm.UpdateMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowRequest(*ssm.UpdateMaintenanceWindowInput) (*request.Request, *ssm.UpdateMaintenanceWindowOutput) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowTarget implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowTarget of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowTarget(*ssm.UpdateMaintenanceWindowTargetInput) (*ssm.UpdateMaintenanceWindowTargetOutput, error) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowTargetRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowTargetRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowTargetRequest(*ssm.UpdateMaintenanceWindowTargetInput) (*request.Request, *ssm.UpdateMaintenanceWindowTargetOutput) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowTargetWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowTargetWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowTargetWithContext(context.Context, *ssm.UpdateMaintenanceWindowTargetInput, ...request.Option) (*ssm.UpdateMaintenanceWindowTargetOutput, error) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowTask implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowTask of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowTask(*ssm.UpdateMaintenanceWindowTaskInput) (*ssm.UpdateMaintenanceWindowTaskOutput, error) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowTaskRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowTaskRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowTaskRequest(*ssm.UpdateMaintenanceWindowTaskInput) (*request.Request, *ssm.UpdateMaintenanceWindowTaskOutput) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowTaskWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowTaskWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowTaskWithContext(context.Context, *ssm.UpdateMaintenanceWindowTaskInput, ...request.Option) (*ssm.UpdateMaintenanceWindowTaskOutput, error) {
	panic("unimplemented")
}

// UpdateMaintenanceWindowWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateMaintenanceWindowWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateMaintenanceWindowWithContext(context.Context, *ssm.UpdateMaintenanceWindowInput, ...request.Option) (*ssm.UpdateMaintenanceWindowOutput, error) {
	panic("unimplemented")
}

// UpdateManagedInstanceRole implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateManagedInstanceRole of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateManagedInstanceRole(*ssm.UpdateManagedInstanceRoleInput) (*ssm.UpdateManagedInstanceRoleOutput, error) {
	panic("unimplemented")
}

// UpdateManagedInstanceRoleRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateManagedInstanceRoleRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateManagedInstanceRoleRequest(*ssm.UpdateManagedInstanceRoleInput) (*request.Request, *ssm.UpdateManagedInstanceRoleOutput) {
	panic("unimplemented")
}

// UpdateManagedInstanceRoleWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateManagedInstanceRoleWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateManagedInstanceRoleWithContext(context.Context, *ssm.UpdateManagedInstanceRoleInput, ...request.Option) (*ssm.UpdateManagedInstanceRoleOutput, error) {
	panic("unimplemented")
}

// UpdateOpsItem implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateOpsItem of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateOpsItem(*ssm.UpdateOpsItemInput) (*ssm.UpdateOpsItemOutput, error) {
	panic("unimplemented")
}

// UpdateOpsItemRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateOpsItemRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateOpsItemRequest(*ssm.UpdateOpsItemInput) (*request.Request, *ssm.UpdateOpsItemOutput) {
	panic("unimplemented")
}

// UpdateOpsItemWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateOpsItemWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateOpsItemWithContext(context.Context, *ssm.UpdateOpsItemInput, ...request.Option) (*ssm.UpdateOpsItemOutput, error) {
	panic("unimplemented")
}

// UpdateOpsMetadata implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateOpsMetadata of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateOpsMetadata(*ssm.UpdateOpsMetadataInput) (*ssm.UpdateOpsMetadataOutput, error) {
	panic("unimplemented")
}

// UpdateOpsMetadataRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateOpsMetadataRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateOpsMetadataRequest(*ssm.UpdateOpsMetadataInput) (*request.Request, *ssm.UpdateOpsMetadataOutput) {
	panic("unimplemented")
}

// UpdateOpsMetadataWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateOpsMetadataWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateOpsMetadataWithContext(context.Context, *ssm.UpdateOpsMetadataInput, ...request.Option) (*ssm.UpdateOpsMetadataOutput, error) {
	panic("unimplemented")
}

// UpdatePatchBaseline implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdatePatchBaseline of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdatePatchBaseline(*ssm.UpdatePatchBaselineInput) (*ssm.UpdatePatchBaselineOutput, error) {
	panic("unimplemented")
}

// UpdatePatchBaselineRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdatePatchBaselineRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdatePatchBaselineRequest(*ssm.UpdatePatchBaselineInput) (*request.Request, *ssm.UpdatePatchBaselineOutput) {
	panic("unimplemented")
}

// UpdatePatchBaselineWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdatePatchBaselineWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdatePatchBaselineWithContext(context.Context, *ssm.UpdatePatchBaselineInput, ...request.Option) (*ssm.UpdatePatchBaselineOutput, error) {
	panic("unimplemented")
}

// UpdateResourceDataSync implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateResourceDataSync of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateResourceDataSync(*ssm.UpdateResourceDataSyncInput) (*ssm.UpdateResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// UpdateResourceDataSyncRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateResourceDataSyncRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateResourceDataSyncRequest(*ssm.UpdateResourceDataSyncInput) (*request.Request, *ssm.UpdateResourceDataSyncOutput) {
	panic("unimplemented")
}

// UpdateResourceDataSyncWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateResourceDataSyncWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateResourceDataSyncWithContext(context.Context, *ssm.UpdateResourceDataSyncInput, ...request.Option) (*ssm.UpdateResourceDataSyncOutput, error) {
	panic("unimplemented")
}

// UpdateServiceSetting implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateServiceSetting of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateServiceSetting(*ssm.UpdateServiceSettingInput) (*ssm.UpdateServiceSettingOutput, error) {
	panic("unimplemented")
}

// UpdateServiceSettingRequest implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateServiceSettingRequest of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateServiceSettingRequest(*ssm.UpdateServiceSettingInput) (*request.Request, *ssm.UpdateServiceSettingOutput) {
	panic("unimplemented")
}

// UpdateServiceSettingWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).UpdateServiceSettingWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) UpdateServiceSettingWithContext(context.Context, *ssm.UpdateServiceSettingInput, ...request.Option) (*ssm.UpdateServiceSettingOutput, error) {
	panic("unimplemented")
}

// WaitUntilCommandExecuted implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).WaitUntilCommandExecuted of mockedSSM.SSMAPI.
func (m *mockedSSM) WaitUntilCommandExecuted(*ssm.GetCommandInvocationInput) error {
	panic("unimplemented")
}

// WaitUntilCommandExecutedWithContext implements ssmiface.SSMAPI.
// Subtle: this method shadows the method (SSMAPI).WaitUntilCommandExecutedWithContext of mockedSSM.SSMAPI.
func (m *mockedSSM) WaitUntilCommandExecutedWithContext(context.Context, *ssm.GetCommandInvocationInput, ...request.WaiterOption) error {
	panic("unimplemented")
}

func (m *mockedSSM) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return m.MockGetParameter(input)
}

type mockedSecretsManager struct {
	secretsmanageriface.SecretsManagerAPI

	MockGetSecretValue func(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

func (m *mockedSecretsManager) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return m.MockGetSecretValue(input)
}

func TestOperators(t *testing.T) {
	cursor := func(s string) *tree.Cursor {
		c, err := tree.ParseCursor(s)
		So(err, ShouldBeNil)
		return c
	}

	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}

	ref := func(s string) *Expr {
		return &Expr{Type: Reference, Reference: cursor(s)}
	}
	str := func(s string) *Expr {
		return &Expr{Type: Literal, Literal: s}
	}
	num := func(v int64) *Expr {
		return &Expr{Type: Literal, Literal: v}
	}
	null := func() *Expr {
		return &Expr{Type: Literal, Literal: nil}
	}
	env := func(s string) *Expr {
		return &Expr{Type: EnvVar, Name: s}
	}
	or := func(l *Expr, r *Expr) *Expr {
		return &Expr{Type: LogicalOr, Left: l, Right: r}
	}

	var exprOk func(*Expr, *Expr)
	exprOk = func(got *Expr, want *Expr) {
		So(got, ShouldNotBeNil)
		So(want, ShouldNotBeNil)

		So(got.Type, ShouldEqual, want.Type)
		switch want.Type {
		case Literal:
			So(got.Literal, ShouldEqual, want.Literal)

		case Reference:
			So(got.Reference.String(), ShouldEqual, want.Reference.String())

		case LogicalOr:
			exprOk(got.Left, want.Left)
			exprOk(got.Right, want.Right)
		}
	}

	Convey("Parser", t, func() {
		Convey("parses op calls in their entirety", func() {
			phase := EvalPhase

			opOk := func(code string, name string, args ...*Expr) {
				op, err := ParseOpcall(phase, code)
				So(err, ShouldBeNil)
				So(op, ShouldNotBeNil)

				_, ok := op.op.(NullOperator)
				So(ok, ShouldBeTrue)
				So(op.op.(NullOperator).Missing, ShouldEqual, name)

				So(len(op.args), ShouldEqual, len(args))
				for i, expect := range args {
					exprOk(op.args[i], expect)
				}
			}

			opErr := func(code string, msg string) {
				_, err := ParseOpcall(phase, code)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, msg)
			}

			opIgnore := func(code string) {
				op, err := ParseOpcall(phase, code)
				So(op, ShouldBeNil)
				So(err, ShouldBeNil)
			}

			Convey("handles opcodes with and without arguments", func() {
				opOk(`(( null 42 ))`, "null", num(42))
				opOk(`(( null 1 2 3 4 ))`, "null", num(1), num(2), num(3), num(4))
			})

			Convey("ignores optional whitespace", func() {
				args := []*Expr{num(1), num(2), num(3)}
				opOk(`((null 1 2 3))`, "null", args...)
				opOk(`((null 1	2	3))`, "null", args...)
				opOk(`((null 1	2	3	))`, "null", args...)
				opOk(`((null 1 	 2 	 3 	 ))`, "null", args...)
			})

			Convey("allows use of commas to separate arguments", func() {
				args := []*Expr{num(1), num(2), num(3)}
				opOk(`((null 1, 2, 3))`, "null", args...)
				opOk(`((null 1,	2,	3))`, "null", args...)
				opOk(`((null 1,	2,	3,	))`, "null", args...)
				opOk(`((null 1 ,	 2 ,	 3 ,	 ))`, "null", args...)
			})

			Convey("allows use of parentheses around arguments", func() {
				args := []*Expr{num(1), num(2), num(3)}
				opOk(`((null(1,2,3)))`, "null", args...)
				opOk(`((null(1, 2, 3) ))`, "null", args...)
				opOk(`((null( 1,	2,	3)))`, "null", args...)
				opOk(`((null (1,	2,	3)	))`, "null", args...)
				opOk(`((null (1 ,	 2 ,	 3)	 ))`, "null", args...)
			})

			Convey("handles string literal arguments", func() {
				opOk(`(( null "string" ))`, "null", str("string"))
				opOk(`(( null "string with whitespace" ))`, "null", str("string with whitespace"))
				opOk(`(( null "a \"quoted\" string" ))`, "null", str(`a "quoted" string`))
				opOk(`(( null "\\escaped" ))`, "null", str(`\escaped`))
			})

			Convey("handles reference (cursor) arguments", func() {
				opOk(`(( null x.y.z ))`, "null", ref("x.y.z"))
				opOk(`(( null x.[0].z ))`, "null", ref("x.0.z"))
				opOk(`(( null x[0].z ))`, "null", ref("x.0.z"))
				opOk(`(( null x[0]z ))`, "null", ref("x.0.z"))
			})

			Convey("handles mixed collections of argument types", func() {
				opOk(`(( xyzzy "string" x.y.z 42  ))`, "xyzzy", str("string"), ref("x.y.z"), num(42))
				opOk(`(( xyzzy("string" x.y.z 42) ))`, "xyzzy", str("string"), ref("x.y.z"), num(42))
			})

			Convey("handles expression-based operands", func() {
				opOk(`(( null meta.key || "default" ))`, "null",
					or(ref("meta.key"), str("default")))

				opOk(`(( null meta.key || "default" "second" ))`, "null",
					or(ref("meta.key"), str("default")),
					str("second"))

				opOk(`(( null meta.key || "default", "second" ))`, "null",
					or(ref("meta.key"), str("default")),
					str("second"))

				opOk(`(( null meta.key || "default", meta.other || nil ))`, "null",
					or(ref("meta.key"), str("default")),
					or(ref("meta.other"), null()))

				opOk(`(( null meta.key || "default"     meta.other || nil ))`, "null",
					or(ref("meta.key"), str("default")),
					or(ref("meta.other"), null()))
			})

			Convey("handles environment variables as operands", func() {
				os.Setenv("SPRUCE_FOO", "first test")
				os.Setenv("_SPRUCE", "_sprucify!")
				os.Setenv("ENOENT", "")
				os.Setenv("http_proxy", "no://thank/you")
				os.Setenv("variable.with.dots", "dots are ok")

				opOk(`(( null $SPRUCE_FOO ))`, "null", env("SPRUCE"))
				opOk(`(( null $_SPRUCE ))`, "null", env("_SPRUCE"))
				opOk(`(( null $ENOENT || $SPRUCE_FOO ))`, "null",
					or(env("ENOENT"), env("SPRUCE_FOO")))
				opOk(`(( null $http_proxy))`, "null", env("http_proxy"))
				opOk(`(( null $variable.with.dots ))`, "null", env("variable.with.dots"))
			})

			Convey("throws errors for malformed expression", func() {
				opErr(`(( null meta.key ||, nil ))`,
					`syntax error near: meta.key ||, nil`)

				opErr(`(( null || ))`,
					`syntax error near: ||`)

				opErr(`(( null || meta.key ))`,
					`syntax error near: || meta.key`)

				opErr(`(( null meta.key || || ))`,
					`syntax error near: meta.key || ||`)
			})

			Convey("ignores spiff-like bang-notation", func() {
				opIgnore(`((!credhub))`)
			})

			Convey("ignores BOSH varnames that aren't null-arity operators", func() {
				opIgnore(`((var-name))`)
			})
		})
	})

	Convey("Expression Engine", t, func() {
		var e *Expr
		var tree map[interface{}]interface{}

		evaluate := func(e *Expr, tree map[interface{}]interface{}) interface{} {
			v, err := e.Evaluate(tree)
			So(err, ShouldBeNil)
			return v
		}

		Convey("Literals evaluate to themselves", func() {
			e = &Expr{Type: Literal, Literal: "value"}
			So(evaluate(e, tree), ShouldEqual, "value")

			e = &Expr{Type: Literal, Literal: ""}
			So(evaluate(e, tree), ShouldEqual, "")

			e = &Expr{Type: Literal, Literal: nil}
			So(evaluate(e, tree), ShouldEqual, nil)
		})

		Convey("References evaluate to the referenced part of the YAML tree", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			e = &Expr{Type: Reference, Reference: cursor("meta.foo")}
			So(evaluate(e, tree), ShouldEqual, "FOO")

			e = &Expr{Type: Reference, Reference: cursor("meta.bar")}
			So(evaluate(e, tree), ShouldEqual, "BAR")
		})

		Convey("|| operator evaluates to the first found value", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			So(evaluate(or(str("first"), str("second")), tree), ShouldEqual, "first")
			So(evaluate(or(ref("meta.foo"), str("second")), tree), ShouldEqual, "FOO")
			So(evaluate(or(ref("meta.ENOENT"), ref("meta.foo")), tree), ShouldEqual, "FOO")
		})

		Convey("|| operator treats nil as a found value", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			So(evaluate(or(null(), str("second")), tree), ShouldBeNil)
			So(evaluate(or(ref("meta.ENOENT"), null()), tree), ShouldBeNil)
		})
	})

	Convey("Expression Reduction Algorithm", t, func() {
		var orig, final *Expr
		var err error

		Convey("ignores singleton expression", func() {
			orig = str("string")
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)

			orig = null()
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)

			orig = ref("meta.key")
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)
		})

		Convey("handles normal alternates that terminated in a literal", func() {
			orig = or(ref("a.b.c"), str("default"))
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)
		})

		Convey("throws errors (warnings) for unreachable alternates", func() {
			orig = or(null(), str("ignored"))
			final, err = orig.Reduce()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `literal nil short-circuits expression (nil || "ignored")`)
			exprOk(final, null())

			orig = or(ref("some.key"), or(str("default"), ref("ignored.key")))
			final, err = orig.Reduce()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `literal "default" short-circuits expression (some.key || "default" || ignored.key)`)
			exprOk(final, or(ref("some.key"), str("default")))

			orig = or(or(ref("some.key"), str("default")), ref("ignored.key"))
			final, err = orig.Reduce()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `literal "default" short-circuits expression (some.key || "default" || ignored.key)`)
			exprOk(final, or(ref("some.key"), str("default")))
		})
	})

	Convey("File Operator", t, func() {
		op := FileOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  sample_file: assets/file_operator/sample.txt
`),
		}
		basedir, _ := os.Getwd()

		Convey("can read a direct file", func() {
			r, err := op.Run(ev, []*Expr{
				str("assets/file_operator/test.txt"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "This is a test\n")
		})

		Convey("can read a file from a reference", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.sample_file"),
			})

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			content, err := os.ReadFile("assets/file_operator/sample.txt")
			So(r.Value.(string), ShouldEqual, string(content))
		})

		Convey("can read a file relative to a specified base path", func() {
			os.Setenv("SPRUCE_FILE_BASE_PATH", filepath.Join(basedir, "assets/file_operator"))
			r, err := op.Run(ev, []*Expr{
				str("test.txt"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "This is a test\n")
		})

		if _, err := os.Stat("/etc/hosts"); err == nil {
			Convey("can read an absolute path", func() {
				os.Setenv("SPRUCE_FILE_BASE_PATH", filepath.Join(basedir, "assets/file_operator"))
				r, err := op.Run(ev, []*Expr{
					str("/etc/hosts"),
				})
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)

				So(r.Type, ShouldEqual, Replace)

				content, err := os.ReadFile("/etc/hosts")
				So(r.Value.(string), ShouldEqual, string(content))
			})
		}

		Convey("can handle a missing file", func() {
			r, err := op.Run(ev, []*Expr{
				str("no_one_should_ever_name_a_file_that_doesnt_exist_this_name"),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

	})

	Convey("Grab Operator", t, func() {
		op := GrabOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`key:
  subkey:
    value: found it
    other: value 2
  list1:
    - first
    - second
  list2:
    - third
    - fourth
  lonely:
    - one
`),
		}

		Convey("can grab a single value", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found it")
		})

		Convey("can grab a single value using an environment variable in the reference", func() {
			os.Setenv("SUB_KEY", "subkey")
			r, err := op.Run(ev, []*Expr{
				ref("key.$SUB_KEY.value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found it")
		})

		Convey("can grab a single list value", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.lonely"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			l, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)

			So(len(l), ShouldEqual, 1)
			So(l[0], ShouldEqual, "one")
		})

		Convey("can grab a multiple lists and flatten them", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.list1"),
				ref("key.lonely"),
				ref("key.list2"),
				ref("key.lonely.0"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			l, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)

			So(len(l), ShouldEqual, 6)
			So(l[0], ShouldEqual, "first")
			So(l[1], ShouldEqual, "second")
			So(l[2], ShouldEqual, "one")
			So(l[3], ShouldEqual, "third")
			So(l[4], ShouldEqual, "fourth")
			So(l[5], ShouldEqual, "one")
		})

		Convey("can grab multiple values", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
				ref("key.subkey.other"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 2)
			So(v[0], ShouldEqual, "found it")
			So(v[1], ShouldEqual, "value 2")
		})

		Convey("flattens constituent arrays", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.list2"),
				ref("key.list1"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 4)
			So(v[0], ShouldEqual, "third")
			So(v[1], ShouldEqual, "fourth")
			So(v[2], ShouldEqual, "first")
			So(v[3], ShouldEqual, "second")
		})

		Convey("throws errors for missing arguments", func() {
			_, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)
		})

		Convey("throws errors for dangling references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.that.does.not.exist"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Environment Variable Resolution (via grab)", t, func() {
		op := GrabOperator{}
		ev := &Evaluator{}
		os.Setenv("GRAB_ONE", "one")
		os.Setenv("GRAB_TWO", "two")
		os.Setenv("GRAB_NOT", "")
		os.Setenv("GRAB_BOOL", "true")
		os.Setenv("GRAB_MULTILINE", `line1

line3
line4`)

		Convey("can grab a single environment value", func() {
			r, err := op.Run(ev, []*Expr{
				env("GRAB_ONE"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "one")
		})

		Convey("tries alternates until it finds a set environment variable", func() {
			r, err := op.Run(ev, []*Expr{
				or(env("GRAB_THREE"), or(env("GRAB_TWO"), env("GRAB_ONE"))),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "two")
		})

		Convey("unmarshalls variable contents", func() {
			r, err := op.Run(ev, []*Expr{
				env("GRAB_BOOL"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(bool), ShouldEqual, true)
		})

		Convey("does not unmarshall string-only variables", func() {
			r, err := op.Run(ev, []*Expr{
				env("GRAB_MULTILINE"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, `line1

line3
line4`)
		})

		Convey("throws errors for unset environment variables", func() {
			_, err := op.Run(ev, []*Expr{
				env("GRAB_NOT"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Concat Operator", t, func() {
		op := ConcatOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`key:
  subkey:
    value: found it
    other: value 2
  list1:
    - first
    - second
  list2:
    - third
    - fourth
douglas:
  adams: 42
math:
  PI: 3.14159
`),
		}

		Convey("can concat a single value", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
				ref("key.list1.0"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found itfirst")
		})

		Convey("can concat a literal values", func() {
			r, err := op.Run(ev, []*Expr{
				str("a literal "),
				str("value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "a literal value")
		})

		Convey("can concat multiple values", func() {
			r, err := op.Run(ev, []*Expr{
				str("I "),
				ref("key.subkey.value"),
				str("!"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "I found it!")
		})

		Convey("can concat integer literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("the answer = "),
				ref("douglas.adams"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "the answer = 42")
		})

		Convey("can concat float literals", func() {
			r, err := op.Run(ev, []*Expr{
				ref("math.PI"),
				str(" is PI"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "3.14159 is PI")
		})

		Convey("throws errors for missing arguments", func() {
			_, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)

			_, err = op.Run(ev, []*Expr{str("one")})
			So(err, ShouldNotBeNil)
		})

		Convey("throws errors for dangling references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.that.does.not.exist"),
				str("string"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("static_ips Operator", t, func() {
		op := StaticIPOperator{}
		Reset(func() {
			UsedIPs = map[string]string{}
		})

		Convey("can resolve valid networks inside of job contexts", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 3)
			So(v[0], ShouldEqual, "10.0.0.5")
			So(v[1], ShouldEqual, "10.0.0.6")
			So(v[2], ShouldEqual, "10.0.0.7")
		})
		Convey("works with new style bosh manifests", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.5 - 10.0.0.10 ]
instance_groups:
- name: job1
  instances: 2
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 2)
			So(v[0], ShouldEqual, "10.0.0.5")
			So(v[1], ShouldEqual, "10.0.0.6")
		})

		Convey("works with multiple subnets", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.3 ]
  - static: [ 10.0.1.5 - 10.0.1.10 ]
instance_groups:
- name: job1
  instances: 4
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2), num(3)})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 4)
			So(v[0], ShouldEqual, "10.0.0.2")
			So(v[1], ShouldEqual, "10.0.0.3")
			So(v[2], ShouldEqual, "10.0.1.5")
			So(v[3], ShouldEqual, "10.0.1.6")
		})

		Convey("works with multiple subnets with an availability zone", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.3 ]
    az: z2
  - static: [ 10.0.1.5 - 10.0.1.10 ]
    az: z1
instance_groups:
- name: job1
  instances: 4
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2), num(3)})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 4)
			So(v[0], ShouldEqual, "10.0.0.2")
			So(v[1], ShouldEqual, "10.0.0.3")
			So(v[2], ShouldEqual, "10.0.1.5")
			So(v[3], ShouldEqual, "10.0.1.6")
		})

		Convey("works with instance_group availability zones", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.3 ]
    az: z1
  - static: [ 10.0.1.5 - 10.0.1.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 3
  azs: [z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 3)
			So(v[0], ShouldEqual, "10.0.1.5")
			So(v[1], ShouldEqual, "10.0.1.6")
			So(v[2], ShouldEqual, "10.0.1.7")
		})

		Convey("works with directly specified availability zones", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.4 ]
    az: z1
  - static: [ 10.0.2.6 - 10.0.2.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 6
  azs: [z1,z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{
				str("z2:1"),
				num(0),
				str("z1:2"),
				str("z2:2"),
				num(1),
				str("z2:4"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 6)
			So(v[0], ShouldEqual, "10.0.2.7")
			So(v[1], ShouldEqual, "10.0.0.2")
			So(v[2], ShouldEqual, "10.0.0.4")
			So(v[3], ShouldEqual, "10.0.2.8")
			So(v[4], ShouldEqual, "10.0.0.3")
			So(v[5], ShouldEqual, "10.0.2.10")
		})

		Convey("throws an error if an unknown availability zone is used in operator", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.4 ]
    az: z1
  - static: [ 10.0.2.6 - 10.0.2.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 2
  azs: [z1,z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{
				str("z2:0"),
				str("z3:1"),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if offset for an availability zone is out of bounds", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.1 - 10.0.0.5 ]
    az: z1
  - static: [ 10.0.2.1 - 10.0.2.5 ]
    az: z2
instance_groups:
- name: job1
  instances: 2
  azs: [z1,z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{
				str("z1:4"),
				str("z1:5"),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if an instance_group availability zone is not found in subnets", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.4 ]
    az: z1
  - static: [ 10.0.2.6 - 10.0.2.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 2
  azs: [z1,z2,z3]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
			}

			r, err := op.Run(ev, []*Expr{
				str("z1:0"),
				str("z2:1"),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("can resolve valid large networks inside of job contexts", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.0 - 10.1.0.1 ]
jobs:
  - name: job1
    instances: 7
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{
				num(0),
				num(255),   // 2^8 - 1
				num(256),   // 2^8
				num(257),   // 2^8 + 1
				num(65535), // 2^16 - 1
				num(65536), // 2^16
				num(65537), // 2^16 + 1

				// 1st octet rollover testing disabled due to improve speed.
				// but verified working on 11/30/2015 - gfranks
				//				num(16777215), // 2^24 - 1
				//				num(16777216), // 2^24
				//				num(16777217), // 2^24 + 1
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 7)
			So(v[0], ShouldEqual, "10.0.0.0")
			So(v[1], ShouldEqual, "10.0.0.255")
			So(v[2], ShouldEqual, "10.0.1.0") //  3rd octet rollover
			So(v[3], ShouldEqual, "10.0.1.1")

			So(v[4], ShouldEqual, "10.0.255.255")
			So(v[5], ShouldEqual, "10.1.0.0") //  2nd octet rollover
			So(v[6], ShouldEqual, "10.1.0.1")

			// 1st octet rollover testing disabled due to improve speed.
			// but verified working on 11/30/2015 - gfranks
			//			So(v[7], ShouldEqual, "10.255.255.255")
			//			So(v[8], ShouldEqual, "11.0.0.0") //  1st octet rollover
			//			So(v[9], ShouldEqual, "11.0.0.1")
		})

		Convey("throws an error if no job name is specified", func() {
			ev := &Evaluator{
				Here: cursor("jobs.0.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if no job instances specified", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if job instances is not a number", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    instances: PI
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if job has no network name", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has no subnets key", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has no subnets", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets: []
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has no static ranges", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
     - {}
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has malformed static range array(s)", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
    - static: [ 10.0.0.1, 10.0.0.254 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network static range has malformed IP addresses", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
    - static: 10.0.0.0.0.0.0.1 - geoff
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if the static address pool is too small", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
    - static: 172.16.31.10 - 172.16.31.11
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if the address pool ends before it starts", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
					`networks:
  - name: test-network
    subnets:
    - static: [ 10.8.0.1 - 10.0.0.255 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
			}

			r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "ends before it starts")
			So(r, ShouldBeNil)
		})
	})

	Convey("inject Operator", t, func() {
		op := InjectOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`key:
  subkey:
    value: found it
    other: value 2
  subkey2:
    value: overridden
    third: trois
  list1:
    - first
    - second
  list2:
    - third
    - fourth
`),
		}

		Convey("can inject a single sub-map", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Inject)
			v, ok := r.Value.(map[interface{}]interface{})
			So(ok, ShouldBeTrue)
			So(v["value"], ShouldEqual, "found it")
			So(v["other"], ShouldEqual, "value 2")
		})

		Convey("can inject multiple sub-maps", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey"),
				ref("key.subkey2"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Inject)
			v, ok := r.Value.(map[interface{}]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 3)
			So(v["value"], ShouldEqual, "overridden")
			So(v["other"], ShouldEqual, "value 2")
			So(v["third"], ShouldEqual, "trois")
		})

		Convey("handles non-existent references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.subkey"),
				ref("key.subkey2"),
				ref("key.subkey2.ENOENT"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error when trying to inject a scalar", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error when trying to inject a list", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.list1"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("param Operator", t, func() {
		op := ParamOperator{}
		ev := &Evaluator{}

		Convey("always causes an error", func() {
			r, err := op.Run(ev, []*Expr{
				str("this is the error"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "this is the error")
			So(r, ShouldBeNil)
		})
	})

	Convey("Join Operator", t, func() {
		op := JoinOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`---
meta:
  authorities:
  - password.write
  - clients.write
  - clients.read
  - scim.write
  - scim.read
  - uaa.admin
  - clients.secret

  secondlist:
  - admin.write
  - admin.read

  emptylist: []

  anotherkey:
  - entry1
  - somekey: value
  - entry2

  somestanza:
    foo: bar
    wom: bat
`),
		}

		Convey("can join a simple list", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.authorities"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret")
		})

		Convey("can join multiple lists", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.authorities"),
				ref("meta.secondlist"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret,admin.write,admin.read")
		})

		Convey("can join an empty list", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.emptylist"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "")
		})

		Convey("can join string literals", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				str("password.write"),
				str("clients.write"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write")
		})

		Convey("can join integer literals", func() {
			r, err := op.Run(ev, []*Expr{
				str(":"),
				num(4), num(8), num(15),
				num(16), num(23), num(42),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "4:8:15:16:23:42")
		})

		Convey("can join referenced string entry", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.somestanza.foo"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "bar")
		})

		Convey("can join referenced string entries", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.somestanza.foo"),
				ref("meta.somestanza.wom"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "bar,bat")
		})

		Convey("can join multiple referenced entries", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.authorities"),
				ref("meta.somestanza.foo"),
				ref("meta.somestanza.wom"),
				str("ending"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret,bar,bat,ending")
		})

		Convey("throws an error when there are no arguments", func() {
			r, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no arguments specified")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when there are too few arguments", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "too few arguments supplied")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when separator argument is not a literal", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.emptylist"),
				ref("meta.authorities"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "join operator only accepts literal argument for the separator")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when referenced entry is not a list or literal", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.somestanza"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "referenced entry is not a list or string")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when referenced list contains non-string entries", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.anotherkey"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "is not compatible for")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when there are unresolvable references", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.non-existent"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Unable to resolve")
			So(r, ShouldBeNil)
		})

		Convey("calculates dependencies correctly", func() {

			//TODO: Move this to a higher scope when more dependencies tests are added
			shouldHaveDeps := func(actual interface{}, expected ...interface{}) string {
				deps := actual.([]*tree.Cursor)
				paths := []string{}
				for _, path := range expected {
					normalizedPath, err := tree.ParseCursor(path.(string))
					if err != nil {
						panic(fmt.Sprintf("improper path %s passed to test", path.(string)))
					}
					paths = append(paths, normalizedPath.String())
				}
				actualPaths := []string{}
				//make an array so we can give some coherent output on error
				for _, dep := range deps {
					//Pass through tree so that tests can tolerate changes to the cursor lib
					actualPaths = append(actualPaths, dep.String())
				}
				//sort and compare
				sort.Strings(actualPaths)
				sort.Strings(paths)
				match := reflect.DeepEqual(actualPaths, paths)
				//give result
				if !match {
					return fmt.Sprintf("actual: %+v\n expected: %+v", actualPaths, paths)
				}
				return ""
			}

			Convey("with a single list", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.secondlist"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.secondlist.[0]", "meta.secondlist.[1]")
			})

			Convey("with multiple lists", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.authorities"),
					ref("meta.secondlist"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.authorities.[0]", "meta.authorities.[1]",
					"meta.authorities.[2]", "meta.authorities.[3]", "meta.authorities.[4]",
					"meta.authorities.[5]", "meta.authorities.[6]",
					"meta.secondlist.[0]", "meta.secondlist.[1]")
			})

			Convey("with a reference string", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.somestanza.foo"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo")
			})

			Convey("with multiple reference strings", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.somestanza.foo"),
					ref("meta.somestanza.wom"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo", "meta.somestanza.wom")
			})

			Convey("with a reference string and a list", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.somestanza.foo"),
					ref("meta.secondlist"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo", "meta.secondlist.[0]",
					"meta.secondlist.[1]")
			})

			Convey("with a literal string", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					str("literally literal"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps)
			})

			Convey("with a literal string and a reference string", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					str("beep"),
					ref("meta.somestanza.foo"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo")
			})
		})
	})

	Convey("empty operator", t, func() {
		op := EmptyOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`---
meta:
  authorities: meep
`),
		}

		//These three are with unquoted arguments (references)
		Convey("can replace with a hash", func() {
			r, err := op.Run(ev, []*Expr{
				ref("hash"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isHash := r.Value.(map[string]interface{})
			So(isHash, ShouldBeTrue)
			So(val, ShouldResemble, map[string]interface{}{})
		})

		Convey("can replace with an array", func() {
			r, err := op.Run(ev, []*Expr{
				ref("array"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isArray := r.Value.([]interface{})
			So(isArray, ShouldBeTrue)
			So(val, ShouldResemble, []interface{}{})
		})

		Convey("can replace with an empty string", func() {
			r, err := op.Run(ev, []*Expr{
				ref("string"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isString := r.Value.(string)
			So(isString, ShouldBeTrue)
			So(val, ShouldEqual, "")
		})

		Convey("throws an error for unrecognized types", func() {
			r, err := op.Run(ev, []*Expr{
				ref("void"),
			})
			So(r, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("works with string literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("hash"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isHash := r.Value.(map[string]interface{})
			So(isHash, ShouldBeTrue)
			So(val, ShouldResemble, map[string]interface{}{})
		})

		Convey("throws an error with no args", func() {
			r, err := op.Run(ev, []*Expr{})
			So(r, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error with too many args", func() {
			r, err := op.Run(ev, []*Expr{
				ref("hash"),
				ref("array"),
			})
			So(r, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

	})

	Convey("ips Operator", t, func() {
		op := IpsOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  base_network: 1.2.3.4/24
  base_ip: 1.2.3.4
  index: 20
  negative_index: -20
  count: 2
`),
		}

		Convey("can build a single IP based on refs (CIDR)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_network"),
				ref("meta.index"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "1.2.3.20")
		})

		Convey("can build a single IP based on refs (IP)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_ip"),
				ref("meta.index"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "1.2.3.24")
		})

		Convey("can build a single IP based on literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("1.2.3.4/24"),
				num(20),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "1.2.3.20")
		})

		Convey("can build a list of IP's based on references (CIDR)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_network"),
				ref("meta.index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.20")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.21")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's based on literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("1.2.3.4/24"),
				num(20),
				num(2),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.20")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.21")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's based on references (IP)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_ip"),
				ref("meta.index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.24")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.25")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's using negative index (IP)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_ip"),
				ref("meta.negative_index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.2.240")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.2.241")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's using negative index (CIDR)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_network"),
				ref("meta.negative_index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.236")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.237")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("bails out if index is outside CIDR size", func() {
			r, err := op.Run(ev, []*Expr{
				str("192.168.1.16/29"),
				num(100),
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Start index 100 exceeds size of subnet 192.168.1.16/29")
			So(r, ShouldBeNil)
		})

		Convey("bails out if count would go outside CIDR size", func() {
			r, err := op.Run(ev, []*Expr{
				str("192.168.1.16/29"),
				num(-1),
				num(3),
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Start index 7 and count 3 would exceed size of subnet 192.168.1.16/29")
			So(r, ShouldBeNil)
		})
	})

	Convey("Base64 Operator", t, func() {
		op := Base64Operator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  sample: "Sample Text To Base64 Encode From Reference"
`),
		}

		Convey("can encode a string literal", func() {
			r, err := op.Run(ev, []*Expr{
				str("Sample Text To Base64 Encode"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZQ==")
		})

		Convey("can encode from a reference", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.sample"),
			})

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			So(r.Value.(string), ShouldEqual, "U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZSBGcm9tIFJlZmVyZW5jZQ==")
		})

		Convey("can handle non string scalar input", func() {
			r, err := op.Run(ev, []*Expr{
				str("one"), num(1),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("can handle non string scalar input (i.e numbers)", func() {
			r, err := op.Run(ev, []*Expr{
				num(1),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

	})

	Convey("Base64Decode Operator", t, func() {
		op := Base64DecodeOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  sample: "U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZSBGcm9tIFJlZmVyZW5jZQ=="
`),
		}

		Convey("can decode from a string literal", func() {
			r, err := op.Run(ev, []*Expr{
				str("U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZQ=="),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "Sample Text To Base64 Encode")
		})

		Convey("can decode from a reference", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.sample"),
			})

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			So(r.Value.(string), ShouldEqual, "Sample Text To Base64 Encode From Reference")
		})
	})

	Convey("awsparam/awssecret operator", t, func() {
		op := AwsOperator{variant: "awsparam"}
		ev := &Evaluator{
			Tree: YAML(`{ "testval": "test", "testmap": {}, "testarr": [] }`),
			Here: &tree.Cursor{},
		}
		mockSSM := &mockedSSM{}
		mockSecretsManager := &mockedSecretsManager{}

		var ssmKey string
		var ssmRet string
		var ssmErr error

		mockSSM.MockGetParameter = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			ssmKey = aws.StringValue(in.Name)
			return &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: aws.String(ssmRet),
				},
			}, ssmErr
		}

		parameterstoreClient = mockSSM
		secretsManagerClient = mockSecretsManager

		Convey("in shared logic", func() {
			Convey("should return error if no key given", func() {
				_, err := op.Run(ev, []*Expr{})
				So(err.Error(), ShouldContainSubstring, "awsparam operator requires at least one argument")
			})

			Convey("should concatenate args", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), num(3)})
				So(err, ShouldBeNil)
				So(ssmKey, ShouldEqual, "123")
			})

			Convey("should resolve references", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testval")})
				So(err, ShouldBeNil)
				So(ssmKey, ShouldEqual, "12test")
			})

			Convey("should not allow references to maps", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testmap")})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.testmap is a map; only scalars are supported here")
			})

			Convey("should not allow references to arrays", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testarr")})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.testarr is a list; only scalars are supported here")
			})

			Convey("without key", func() {
				ssmRet = "testx"
				r, err := op.Run(ev, []*Expr{str("val1")})
				So(err, ShouldBeNil)
				So(r.Type, ShouldEqual, Replace)
				So(r.Value.(string), ShouldEqual, "testx")
			})

			Convey("with key", func() {
				Convey("should parse subkey and extract if provided", func() {
					ssmRet = `{ "key": "val" }`
					r, err := op.Run(ev, []*Expr{str("val2?key=key")})
					So(err, ShouldBeNil)
					So(r.Type, ShouldEqual, Replace)
					So(r.Value.(string), ShouldEqual, "val")
				})

				Convey("should error if document not valid yaml / json", func() {
					ssmRet = `key: {`
					_, err := op.Run(ev, []*Expr{str("val3?key=key")})
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "$.val3 error extracting key: yaml: line 1: did not find expected node content")
				})

				Convey("should error if subkey invalid", func() {
					ssmRet = `key: {}`
					_, err := op.Run(ev, []*Expr{str("val4?key=noexist")})
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "$.val4 invalid key 'noexist'")
				})
			})

			Convey("should not call AWS API if SkipAws true", func() {
				SkipAws = true
				count := 0
				mockSSM.MockGetParameter = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					count = count + 1
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(""),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("skipaws")})
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
				SkipAws = false
			})
		})

		Convey("awsparam", func() {
			Convey("should cache lookups", func() {
				count := 0
				mockSSM.MockGetParameter = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					count = count + 1
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(""),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)
				_, err = op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)

				So(count, ShouldEqual, 1)
			})
		})

		Convey("awssecret", func() {
			op = AwsOperator{variant: "awssecret"}
			Convey("should cache lookups", func() {
				count := 0
				mockSecretsManager.MockGetSecretValue = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
					count = count + 1
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)
				_, err = op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)

				So(count, ShouldEqual, 1)
			})

			Convey("should use stage if provided", func() {
				stage := ""
				mockSecretsManager.MockGetSecretValue = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
					stage = aws.StringValue(in.VersionStage)
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val6?stage=test")})
				So(err, ShouldBeNil)

				So(stage, ShouldEqual, "test")
			})

			Convey("should use version if provided", func() {
				version := ""
				mockSecretsManager.MockGetSecretValue = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
					version = aws.StringValue(in.VersionId)
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val7?version=test")})
				So(err, ShouldBeNil)

				So(version, ShouldEqual, "test")
			})
		})
	})

	Convey("Stringify Operator", t, func() {
		op := StringifyOperator{}
		ev := &Evaluator{
			Tree: YAML(`meta:
  map:
    bar: foo
    foo: bar
  list:
  - first
  - second
  scalars:
    bool: true
    number: 42
    string: foobar
`),
		}

		Convey("cannot use operator with more than one reference", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.map"),
				ref("list.0"),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("can stringify map", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.map"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, `bar: foo
foo: bar
`)
		})

		Convey("can stringify list", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.list"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, `- first
- second
`)
		})

		Convey("can stringify scalars", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.scalars"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, `bool: true
number: 42
string: foobar
`)
		})

		Convey("retain string literal", func() {
			r, err := op.Run(ev, []*Expr{
				str("foo"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "foo")
		})

		Convey("retain null literal", func() {
			r, err := op.Run(ev, []*Expr{
				null(),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value, ShouldBeNil)
		})

		Convey("throws errors for dangling references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.that.does.not.exist"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws errors for missing arguments", func() {
			_, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)
		})
	})
}
