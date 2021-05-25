using System.Linq;
using Xunit;

namespace Kiota.Builder.Refiners.Tests {
    public class CSharpLanguageRefinerTests {
        private readonly CodeNamespace root = CodeNamespace.InitRootNamespace();
        #region CommonLanguageRefinerTests
        [Fact]
        public void ConvertsUnionTypesToWrapper() {
            var model = root.AddClass(new CodeClass (root) {
                Name = "model",
                ClassKind = CodeClassKind.Model
            }).First();
            var union = new CodeUnionType(model) {
                Name = "union",
            };
            union.AddType(new (model) {
                Name = "type1",
            }, new(model) {
                Name = "type2"
            });
            var property = model.AddProperty(new CodeProperty(model) {
                Name = "deserialize",
                PropertyKind = CodePropertyKind.Deserializer,
                Type = union,
            }).First();
            var method = model.AddMethod(new CodeMethod(model) {
                Name = "method",
                ReturnType = union
            }).First();
            var parameter = new CodeParameter(method) {
                Name = "param1",
                Type = union
            };
            var indexer = new CodeIndexer(model) {
                Name = "idx",
                ReturnType = union,
            };
            model.SetIndexer(indexer);
            method.AddParameter(parameter);
            ILanguageRefiner.Refine(GenerationLanguage.CSharp, root); //using CSharp so the indexer doesn't get removed
            Assert.True(property.Type is CodeType);
            Assert.True(parameter.Type is CodeType);
            Assert.True(method.ReturnType is CodeType);
            Assert.True(indexer.ReturnType is CodeType);
        }
        [Fact]
        public void MovesClassesWithNamespaceNamesUnderNamespace() {
            var graphNS = root.AddNamespace("graph");
            var modelNS = graphNS.AddNamespace("graph.model");
            var model = graphNS.AddClass(new CodeClass (graphNS) {
                Name = "model",
                ClassKind = CodeClassKind.Model
            }).First();
            ILanguageRefiner.Refine(GenerationLanguage.CSharp, root);
            Assert.Single(root.GetChildElements(true));
            Assert.Single(graphNS.GetChildElements(true));
            Assert.Single(modelNS.GetChildElements(true));
            Assert.Equal(modelNS, model.Parent);
        }
        #endregion
    }
}
