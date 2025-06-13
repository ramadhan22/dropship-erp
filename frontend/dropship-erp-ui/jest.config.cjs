// jest.config.cjs
module.exports = {
  testEnvironment: "jsdom",
  moduleFileExtensions: ["ts", "tsx", "js"],
  transform: {
    "^.+\\.(ts|tsx)$": ["ts-jest", { useESM: true }]
  },
  extensionsToTreatAsEsm: [".ts", ".tsx"],
  globals: {
    "ts-jest": {
      useESM: true,
      tsconfig: "./tsconfig.json",
      diagnostics: false,
    },
  },
  setupFilesAfterEnv: ["<rootDir>/src/setupTests.ts"]
};