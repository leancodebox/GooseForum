import { readFile, readdir } from "node:fs/promises";
import { dirname, extname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";
import { parseBackendMessageCodes } from "./check-server-message-i18n.mjs";

const root = fileURLToPath(new URL("../../", import.meta.url));

export function inspectDictionary(source, file = "dictionary.ts") {
  const ast = extname(file) === ".json"
    ? ts.parseJsonText(file, source)
    : ts.createSourceFile(file, source, ts.ScriptTarget.Latest, true);
  const errors = [];
  let exported;
  function visit(node) {
    if (ts.isObjectLiteralExpression(node)) {
      const seen = new Set();
      for (const property of node.properties) {
        if (!property.name || ts.isComputedPropertyName(property.name)) continue;
        const key = property.name.text;
        if (seen.has(key)) errors.push(`重复 key: ${key}`);
        seen.add(key);
      }
    }
    if (ts.isExportAssignment(node)) {
      let expression = node.expression;
      while (ts.isAsExpression(expression) || ts.isSatisfiesExpression(expression) || ts.isParenthesizedExpression(expression)) expression = expression.expression;
      if (ts.isObjectLiteralExpression(expression)) exported = expression;
    }
    ts.forEachChild(node, visit);
  }
  visit(ast);
  const strings = new Map();
  if (exported) {
    for (const property of exported.properties) {
      if (ts.isPropertyAssignment(property) && ts.isStringLiteral(property.initializer)) strings.set(property.name.text, property.initializer.text);
    }
  }
  const englishImport = ast.statements.find(node => ts.isImportDeclaration(node) && node.importClause?.name?.text === "en");
  const inheritsEnglish = exported?.properties.some(property => ts.isSpreadAssignment(property) && property.expression.getText(ast) === "en");
  return { errors, strings, englishFile: inheritsEnglish && englishImport ? englishImport.moduleSpecifier.text : undefined };
}

async function filesIn(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  return (await Promise.all(entries.map(entry => entry.isDirectory() ? filesIn(resolve(directory, entry.name)) : resolve(directory, entry.name)))).flat();
}

async function main() {
  const files = (await Promise.all([
    "resource/packages/client/src/i18n/messages",
    "resource/apps/web/src/admin/messages",
    "app/bundles/i18n/locales",
  ].map(directory => filesIn(resolve(root, directory))))).flat().filter(file => [".ts", ".json"].includes(extname(file)));
  const dictionaries = new Map(await Promise.all(files.map(async file => [file, inspectDictionary(await readFile(file, "utf8"), file)])));
  const codes = new Set(parseBackendMessageCodes(await readFile(resolve(root, "app/http/controllers/component/message_code.go"), "utf8")));
  const errors = [];
  for (const [file, dictionary] of dictionaries) {
    const label = file.slice(root.length);
    for (const error of dictionary.errors) errors.push(`${label}: ${error}`);
    if (dictionary.englishFile) {
      let englishFile = resolve(dirname(file), dictionary.englishFile).replace(/\.js$/, ".ts");
      if (!extname(englishFile)) englishFile += ".ts";
      const english = dictionaries.get(englishFile);
      if (!english) errors.push(`${label}: 英文回退词典不存在`);
      else for (const [key, value] of dictionary.strings) {
        if (english.strings.get(key) === value) errors.push(`${label}: ${key} 重复声明英文回退值`);
      }
    }
    if (extname(file) === ".json") {
      const data = JSON.parse(await readFile(file, "utf8"));
      for (const key of Object.keys(data)) {
        if (codes.has(key)) errors.push(`${label}: API 消息码 ${key} 应由前端 server-messages 翻译`);
      }
    }
  }
  if (errors.length) {
    console.error(errors.join("\n"));
    process.exitCode = 1;
  } else console.log(`Dictionary audit passed: ${files.length} files; no duplicate keys, redundant fallback strings or misplaced API messages.`);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();
