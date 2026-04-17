package v2

import (
	"strconv"
	"strings"

	sharedValidation "github.com/pdok/smooth-operator/pkg/validation"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/set"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (wmts *WMTS) ValidateCreate(c client.Client) ([]string, error) {
	_ = c
	warnings := []string{}
	allErrs := field.ErrorList{}
	AddGeneralValidationErrorsAndWarnings(wmts, &warnings, &allErrs)

	if len(allErrs) == 0 {
		return warnings, nil
	}
	return warnings, apierrors.NewInvalid(
		GroupKind,
		wmts.GetName(), allErrs)
}

func (wmts *WMTS) ValidateUpdate(c client.Client, wmtsOld *WMTS) ([]string, error) {
	_ = c
	warnings := []string{}
	allErrs := field.ErrorList{}
	AddGeneralValidationErrorsAndWarnings(wmts, &warnings, &allErrs)

	checkChangedUrls(wmts, wmtsOld, &warnings, &allErrs)
	sharedValidation.ValidateLabelsOnUpdate(wmtsOld.Labels, wmts.Labels, &allErrs)

	if len(allErrs) == 0 {
		return warnings, nil
	}

	return warnings, apierrors.NewInvalid(
		GroupKind,
		wmts.GetName(), allErrs)
}

// AddGeneralValidationErrorsAndWarnings Validates the WMTS as an independent unit and adds warnings and errors to lists
func AddGeneralValidationErrorsAndWarnings(wmts *WMTS, warnings *[]string, allErrs *field.ErrorList) {
	if strings.Contains(wmts.GetName(), "wmts") {
		sharedValidation.AddWarning(warnings, *field.NewPath("metadata").Child("name"), "name should not contain wmts", wmts.GroupVersionKind(), wmts.GetName())
	}

	layerIDSet := set.Set[string]{}
	layersPath := field.NewPath("spec").Child("service").Child("layers")
	for index, layer := range wmts.Spec.Service.Layers {
		if layerIDSet.Has(layer.Identifier) {
			*allErrs = append(*allErrs, field.Duplicate(layersPath.Index(index).Child("identifier"), layer.Identifier))
		}
		layerIDSet.Insert(layer.Identifier)
	}

	checkZoomLevels(wmts, warnings, allErrs)

	err := sharedValidation.ValidateIngressRouteURLsContainsBaseURL(wmts.Spec.IngressRouteURLs, wmts.Spec.Service.BaseURL, nil)
	if err != nil {
		*allErrs = append(*allErrs, err)
	}

}

func checkZoomLevels(wmts *WMTS, _ *[]string, allErrs *field.ErrorList) {
	tileMatrixSetPath := field.NewPath("spec").Child("service").Child("tileMatrixSets")
	for index, tileMatrixSet := range wmts.Spec.Service.TileMatrixSets {

		intranges := make([]intRange, 0)
		for _, zoomLevel := range tileMatrixSet.ZoomLevels {
			var zoomLevelRange intRange
			asInt, err := strconv.Atoi(zoomLevel)
			if err == nil {
				zoomLevelRange = intRange{
					minval: asInt,
					maxval: asInt,
				}
			} else {
				parts := strings.Split(zoomLevel, "-")
				if len(parts) != 2 {
					*allErrs = append(*allErrs, field.Invalid(tileMatrixSetPath.Index(index), zoomLevel, "Invalid value for zoomlevel (should be impossible)"))
					continue
				}
				rangeStart, err := strconv.Atoi(parts[0])
				if err != nil {
					*allErrs = append(*allErrs, field.Invalid(tileMatrixSetPath.Index(index), zoomLevel, "Invalid value for zoomlevel (should be impossible)"))
					continue
				}

				rangeEnd, err := strconv.Atoi(parts[0])
				if err != nil {
					*allErrs = append(*allErrs, field.Invalid(tileMatrixSetPath.Index(index), zoomLevel, "Invalid value for zoomlevel (should be impossible)"))
					continue
				}

				if rangeEnd < rangeStart {
					*allErrs = append(*allErrs, field.Invalid(tileMatrixSetPath.Index(index), zoomLevel, "Range end must not be smaller than range start"))
					continue
				}
				zoomLevelRange = intRange{
					minval: rangeStart,
					maxval: rangeEnd,
				}
			}
			if intRangeOverlapsOtherIntRange(zoomLevelRange, intranges) {
				*allErrs = append(*allErrs, field.Invalid(tileMatrixSetPath.Index(index), zoomLevel, "Zoom level overlaps with other zoom level"))
				continue
			}
			intranges = append(intranges, zoomLevelRange)

		}
	}
}

type intRange struct {
	minval int
	maxval int
}

func intRangeOverlapsOtherIntRange(current intRange, others []intRange) bool {
	if len(others) == 0 {
		return false
	}

	for _, other := range others {
		if current.maxval < other.minval {
			continue
		}

		if current.minval > other.maxval {
			continue
		}

		return true
	}

	return false
}

func checkChangedUrls(wmtsNew *WMTS, wmtsOld *WMTS, _ *[]string, allErrs *field.ErrorList) {
	sharedValidation.ValidateIngressRouteURLsNotRemoved(wmtsOld.Spec.IngressRouteURLs, wmtsNew.Spec.IngressRouteURLs, allErrs, nil)

	if len(wmtsNew.Spec.IngressRouteURLs) == 0 {
		// There are no ingressRouteURLs given, spec.service.url is immutable is that case.
		path := field.NewPath("spec").Child("service").Child("url")
		sharedValidation.CheckURLImmutability(
			wmtsOld.Spec.Service.BaseURL,
			wmtsNew.Spec.Service.BaseURL,
			allErrs,
			path,
		)
	} else if wmtsOld.Spec.Service.BaseURL.String() != wmtsNew.Spec.Service.BaseURL.String() {
		// Make sure both the old spec.service.url and the new one are included in the ingressRouteURLs list.
		err := sharedValidation.ValidateIngressRouteURLsContainsBaseURL(wmtsNew.Spec.IngressRouteURLs, wmtsOld.Spec.Service.BaseURL, nil)
		if err != nil {
			*allErrs = append(*allErrs, err)
		}

	}
}
