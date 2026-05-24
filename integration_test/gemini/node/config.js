class TestConfig {
  constructor() {
    this.apiKey = process.env.TEST_AXONHUB_API_KEY || '';
    this.baseUrl = process.env.TEST_GEMINI_BASE_URL || 'http://localhost:8090/gemini';
    this.model = process.env.TEST_MODEL || 'gemini-2.5-flash';
    this.timeout = 30000;
    this.maxRetries = 3;
  }

  validateConfig() {
    if (!this.apiKey) {
      throw new Error('API key is required (set TEST_AXONHUB_API_KEY environment variable)');
    }

    if (!this.model) {
      throw new Error('model is required (set TEST_MODEL environment variable)');
    }
  }

  getHeaders() {
    return {};
  }
}

module.exports = { TestConfig };
