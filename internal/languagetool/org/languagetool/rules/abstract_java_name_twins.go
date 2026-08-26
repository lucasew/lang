
package rules

// Java abstract-class name twins (implementation types use shorter Go names).

// AbstractCheckCaseRule is constructed via NewAbstractCheckCaseRule → *AbstractSimpleReplaceRule2.
type AbstractCheckCaseRule = AbstractSimpleReplaceRule2

// AbstractFutureDateFilter is the core future-date filter.
type AbstractFutureDateFilter = FutureDateFilterCore

// AbstractMakeContractionsFilter aliases MakeContractionsFilter.
type AbstractMakeContractionsFilter = MakeContractionsFilter

// AbstractNewYearDateFilter aliases NewYearDateFilterCore.
type AbstractNewYearDateFilter = NewYearDateFilterCore

// AbstractNumberInWordFilter aliases NumberInWordFilter.
type AbstractNumberInWordFilter = NumberInWordFilter

// AbstractTextToNumberFilter aliases TextToNumberFilter.
type AbstractTextToNumberFilter = TextToNumberFilter
