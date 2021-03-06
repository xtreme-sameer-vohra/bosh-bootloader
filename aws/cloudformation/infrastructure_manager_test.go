package cloudformation_test

import (
	"errors"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InfrastructureManager", func() {
	var (
		builder               *fakes.TemplateBuilder
		stackManager          *fakes.StackManager
		cloudFormationClient  *fakes.CloudFormationClient
		infrastructureManager cloudformation.InfrastructureManager
	)

	BeforeEach(func() {
		cloudFormationClient = &fakes.CloudFormationClient{}

		builder = &fakes.TemplateBuilder{}
		builder.BuildCall.Returns.Template = templates.Template{
			AWSTemplateFormatVersion: "some-template-version",
			Description:              "some-description",
		}

		stackManager = &fakes.StackManager{}

		infrastructureManager = cloudformation.NewInfrastructureManager(builder, stackManager)
	})

	Describe("Create", func() {
		BeforeEach(func() {
			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{Name: "some-stack-name"}
		})

		It("creates the underlying infrastructure and returns the stack", func() {
			stack, err := infrastructureManager.Create("some-key-pair-name", 2, "some-stack-name", "some-lb-type", cloudFormationClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(stack).To(Equal(cloudformation.Stack{Name: "some-stack-name"}))
			Expect(builder.BuildCall.Receives.KeyPairName).To(Equal("some-key-pair-name"))
			Expect(builder.BuildCall.Receives.NumberOfAZs).To(Equal(2))
			Expect(builder.BuildCall.Receives.LBType).To(Equal("some-lb-type"))

			Expect(stackManager.CreateOrUpdateCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.CreateOrUpdateCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.CreateOrUpdateCall.Receives.Template).To(Equal(templates.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
			}))

			Expect(stackManager.WaitForCompletionCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.WaitForCompletionCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(15 * time.Second))
			Expect(stackManager.WaitForCompletionCall.Receives.Action).To(Equal("applying cloudformation template"))

			Expect(stackManager.DescribeCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))
		})

		Context("failure cases", func() {
			It("returns an error when stack can't be created or updated", func() {
				stackManager.CreateOrUpdateCall.Returns.Error = errors.New("stack create or update failed")

				_, err := infrastructureManager.Create("some-key-pair-name", 0, "some-stack-name", "", cloudFormationClient)
				Expect(err).To(MatchError("stack create or update failed"))
			})

			It("returns an error when waiting for stack completion fails", func() {
				stackManager.WaitForCompletionCall.Returns.Error = errors.New("stack wait for completion failed")

				_, err := infrastructureManager.Create("some-key-pair-name", 0, "some-stack-name", "", cloudFormationClient)
				Expect(err).To(MatchError("stack wait for completion failed"))
			})

			It("returns an error when describing the stack fails", func() {
				stackManager.DescribeCall.Returns.Error = errors.New("stack describe failed")

				_, err := infrastructureManager.Create("some-key-pair-name", 0, "some-stack-name", "", cloudFormationClient)
				Expect(err).To(MatchError("stack describe failed"))
			})
		})
	})

	Describe("Exists", func() {
		It("returns true when the stack exists", func() {
			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{}

			exists, err := infrastructureManager.Exists("some-stack-name", cloudFormationClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("returns false when the stack does not exist", func() {
			stackManager.DescribeCall.Returns.Error = cloudformation.StackNotFound

			exists, err := infrastructureManager.Exists("some-stack-name", cloudFormationClient)

			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		Describe("failure cases", func() {
			It("returns an error when the stack manager returns a different error", func() {
				stackManager.DescribeCall.Returns.Error = errors.New("some other error")
				_, err := infrastructureManager.Exists("some-stack-name", cloudFormationClient)
				Expect(err).To(MatchError("some other error"))
			})
		})
	})

	Describe("Delete", func() {
		It("deletes the underlying infrastructure", func() {
			err := infrastructureManager.Delete(cloudFormationClient, "some-stack-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(stackManager.DeleteCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.DeleteCall.Receives.StackName).To(Equal("some-stack-name"))

			Expect(stackManager.WaitForCompletionCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.WaitForCompletionCall.Receives.StackName).To(Equal("some-stack-name"))
			Expect(stackManager.WaitForCompletionCall.Receives.SleepInterval).To(Equal(15 * time.Second))
			Expect(stackManager.WaitForCompletionCall.Receives.Action).To(Equal("deleting cloudformation stack"))
		})

		Context("when the stack goes away after being deleted", func() {
			It("returns without an error", func() {
				stackManager.WaitForCompletionCall.Returns.Error = cloudformation.StackNotFound

				err := infrastructureManager.Delete(cloudFormationClient, "some-stack-name")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("failure cases", func() {
			Context("when the stack fails to delete", func() {
				It("returns an error", func() {
					stackManager.DeleteCall.Returns.Error = errors.New("failed to delete stack")

					err := infrastructureManager.Delete(cloudFormationClient, "some-stack-name")
					Expect(err).To(MatchError("failed to delete stack"))
				})
			})
			Context("when the waiting for completion fails", func() {
				It("returns an error", func() {
					stackManager.WaitForCompletionCall.Returns.Error = errors.New("wait for completion failed")

					err := infrastructureManager.Delete(cloudFormationClient, "some-stack-name")
					Expect(err).To(MatchError("wait for completion failed"))
				})
			})
		})
	})
})
