package v2

//nolint:revive // Complains about the dot imports
import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pdok/smooth-operator/pkg/validation"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	smoothoperatormodel "github.com/pdok/smooth-operator/model"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("WMTS Webhook", func() {
	var (
		obj       *pdoknlv2.WMTS
		oldObj    *pdoknlv2.WMTS
		validator WMTSCustomValidator
	)

	BeforeEach(func() {
		validator = WMTSCustomValidator{k8sClient}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")

		sample := &pdoknlv2.WMTS{}
		err := readSample(sample)
		Expect(err).ToNot(HaveOccurred(), "Reading and parsing the WMTS V2 sample failed")

		obj = sample.DeepCopy()
		oldObj = sample.DeepCopy()

		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
	})

	AfterEach(func() {

	})

	Context("When creating or updating WMTS under Validating Webhook", func() {
		ctx := context.Background()

		It("Creates the WMTS from the sample", func() {
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny creation if there are no labels", func() {
			obj.Labels = nil
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Required(
				field.NewPath("metadata").Child("labels"),
				"can't be empty",
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny Create when URL not in IngressRouteURLs", func() {
			url, err := smoothoperatormodel.ParseURL("http://changed/changed")
			Expect(err).ToNot(HaveOccurred())
			obj.Spec.IngressRouteURLs = []smoothoperatormodel.IngressRouteURL{{URL: smoothoperatormodel.URL{URL: url}}}
			url, err = smoothoperatormodel.ParseURL("http://sample/sample")
			Expect(err).ToNot(HaveOccurred())
			obj.Spec.Service.BaseURL = smoothoperatormodel.URL{URL: url}

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Invalid(
				field.NewPath("spec").Child("ingressRouteUrls"),
				fmt.Sprint(obj.Spec.IngressRouteURLs),
				fmt.Sprintf("must contain baseURL: %s", url),
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Warns if the name contains WMTS", func() {
			obj.Name += "-wmts"
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(Equal(getValidationWarnings(
				obj,
				*field.NewPath("metadata").Child("name"),
				"name should not contain wmts",
				[]string{},
			)))
		})

		It("Should deny Create when minReplicas are larger than maxReplicas", func() {
			obj.Spec.HorizontalPodAutoscalerPatch = &pdoknlv2.HorizontalPodAutoscalerPatch{
				MinReplicas: ptr.To(int32(10)),
				MaxReplicas: ptr.To(int32(5)),
			}

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Invalid(
				field.NewPath("spec").Child("horizontalPodAutoscaler"),
				fmt.Sprintf("minReplicas: %d, maxReplicas: %d", 10, 5),
				"maxReplicas cannot be less than minReplicas",
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny Create when mapserver container doesn't have ephemeral storage", func() {
			obj.Spec.PodSpecPatch = corev1.PodSpec{}

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Required(field.NewPath("spec").
				Child("podSpecPatch").
				Child("containers").
				Key("mapserver").
				Child("resources").
				Child("limits").
				Child(corev1.ResourceEphemeralStorage.String()), ""))))
			Expect(warnings).To(BeEmpty())
		})

		//nolint:gocritic
		//It("Should deny creation if multiple featureTypes have the same name", func() {
		//	Expect(len(obj.Spec.Service.FeatureTypes)).To(BeNumerically(">", 1))
		//	obj.Spec.Service.FeatureTypes[1].Name = obj.Spec.Service.FeatureTypes[0].Name
		//	warnings, err := validator.ValidateCreate(ctx, obj)
		//	Expect(err).To(Equal(getValidationError(obj, field.Duplicate(
		//		field.NewPath("spec").Child("service").Child("featureTypes").Index(1).Child("name"),
		//		obj.Spec.Service.FeatureTypes[1].Name,
		//	))))
		//	Expect(warnings).To(BeEmpty())
		//})

		It("Should deny update if a ingressRouteURL was removed", func() {
			url, err := smoothoperatormodel.ParseURL("http://new.url/path")
			Expect(err).ToNot(HaveOccurred())
			oldObj.Spec.IngressRouteURLs = []smoothoperatormodel.IngressRouteURL{
				{URL: obj.Spec.Service.BaseURL},
				{URL: smoothoperatormodel.URL{URL: url}},
			}
			obj.Spec.IngressRouteURLs = []smoothoperatormodel.IngressRouteURL{{URL: obj.Spec.Service.BaseURL}}
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Invalid(
				field.NewPath("spec").Child("ingressRouteUrls"),
				fmt.Sprint(obj.Spec.IngressRouteURLs),
				fmt.Sprintf("urls cannot be removed, missing: %s", smoothoperatormodel.IngressRouteURL{URL: smoothoperatormodel.URL{URL: url}}),
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Should accept update if a url was changed when it's in ingressRouteUrls", func() {
			url, err := smoothoperatormodel.ParseURL("http://new.url/path")
			Expect(err).ToNot(HaveOccurred())
			oldObj.Spec.IngressRouteURLs = []smoothoperatormodel.IngressRouteURL{
				{URL: obj.Spec.Service.BaseURL},
				{URL: smoothoperatormodel.URL{URL: url}},
			}
			obj.Spec.IngressRouteURLs = oldObj.Spec.IngressRouteURLs
			oldObj.Spec.Service.BaseURL = obj.Spec.Service.BaseURL
			obj.Spec.Service.BaseURL = smoothoperatormodel.URL{URL: url}

			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).ToNot(HaveOccurred())
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny update if a url was changed and ingressRouteUrls = nil", func() {
			url, err := smoothoperatormodel.ParseURL("http://new.url/path")
			Expect(err).ToNot(HaveOccurred())
			obj.Spec.Service.BaseURL = smoothoperatormodel.URL{URL: url}
			obj.Spec.IngressRouteURLs = nil
			oldObj.Spec.IngressRouteURLs = nil

			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Forbidden(
				field.NewPath("spec").Child("service").Child("url"),
				"is immutable, add the old and new urls to spec.ingressRouteUrls in order to change this field",
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny update url was changed but not added to ingressRouteURLs", func() {
			url, err := smoothoperatormodel.ParseURL("http://new.url/path")
			Expect(err).ToNot(HaveOccurred())
			oldObj.Spec.IngressRouteURLs = nil
			obj.Spec.IngressRouteURLs = []smoothoperatormodel.IngressRouteURL{{URL: oldObj.Spec.Service.BaseURL}}
			obj.Spec.Service.BaseURL = smoothoperatormodel.URL{URL: url}
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Invalid(
				field.NewPath("spec").Child("ingressRouteUrls"),
				fmt.Sprint(obj.Spec.IngressRouteURLs),
				fmt.Sprintf("must contain baseURL: %s", obj.Spec.Service.BaseURL),
			))))
			Expect(warnings).To(BeEmpty())

			obj.Spec.IngressRouteURLs = []smoothoperatormodel.IngressRouteURL{{URL: smoothoperatormodel.URL{URL: url}}}
			warnings, err = validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Invalid(
				field.NewPath("spec").Child("ingressRouteUrls"),
				fmt.Sprint(obj.Spec.IngressRouteURLs),
				fmt.Sprintf("must contain baseURL: %s", oldObj.Spec.Service.BaseURL),
			))))
			Expect(warnings).To(BeEmpty())

		})

		It("Should deny update if a label was removed", func() {
			oldKey := ""
			for label := range obj.Labels {
				oldKey = label
				delete(obj.Labels, label)
				break
			}
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Required(
				field.NewPath("metadata").Child("labels").Child(oldKey),
				"labels cannot be removed",
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny update if a label changed", func() {
			oldKey := ""
			oldValue := ""
			newValue := ""
			for label, val := range obj.Labels {
				oldKey = label
				oldValue = val
				newValue = val + "-newval"
				obj.Labels[label] = newValue
				break
			}
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Invalid(
				field.NewPath("metadata").Child("labels").Child(oldKey),
				newValue,
				"immutable: should be: "+oldValue,
			))))
			Expect(warnings).To(BeEmpty())
		})

		It("Should deny update if a label was added", func() {
			newKey := "new-label"
			obj.Labels[newKey] = "test"
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(Equal(getValidationError(obj, field.Forbidden(
				field.NewPath("metadata").Child("labels").Child(newKey),
				"new labels cannot be added",
			))))
			Expect(warnings).To(BeEmpty())
		})

	})

})

func readSample(wmts *pdoknlv2.WMTS) error {
	sampleYaml, err := os.ReadFile("test_data/v3_wmts.yaml")
	if err != nil {
		return err
	}
	sampleJSON, err := yaml.YAMLToJSONStrict(sampleYaml)
	if err != nil {
		return err
	}
	err = json.Unmarshal(sampleJSON, wmts)
	if err != nil {
		return err
	}

	return nil
}

func getValidationError(obj *pdoknlv2.WMTS, errorList *field.Error) error {
	return apierrors.NewInvalid(pdoknlv2.GroupKind, obj.GetName(), field.ErrorList{errorList})
}

func getValidationWarnings(obj *pdoknlv2.WMTS, path field.Path, warning string, warnings []string) admission.Warnings {
	validation.AddWarning(&warnings, path, warning, schema.GroupVersionKind{
		Group:   pdoknlv2.GroupKind.String(),
		Version: "v3",
		Kind:    "WMTS",
	}, obj.GetName())
	return warnings
}
