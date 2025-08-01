import React, { useState, useEffect, useCallback, useRef } from 'react';

// --- ICONS (using inline SVGs for simplicity) ---
const ChartBarIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400">
    <path d="M3 3v18h18" /><path d="M9 17V9" /><path d="M15 17V5" /><path d="M12 17V13" />
  </svg>
);

const ListIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400">
    <line x1="8" x2="21" y1="6" y2="6" /><line x1="8" x2="21" y1="12" y2="12" /><line x1="8" x2="21" y1="18" y2="18" /><line x1="3" x2="3.01" y1="6" y2="6" /><line x1="3" x2="3.01" y1="12" y2="12" /><line x1="3" x2="3.01" y1="18" y2="18" />
  </svg>
);

const ChevronDownIcon = ({ className }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="m6 9 6 6 6-6" />
  </svg>
);

const ServerIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400"><rect width="20" height="8" x="2" y="2" rx="2" ry="2"></rect><rect width="20" height="8" x="2" y="14" rx="2" ry="2"></rect><line x1="6" x2="6.01" y1="6" y2="6"></line><line x1="6" x2="6.01" y1="18" y2="18"></line></svg>
);

const ZapIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
);

const AlertTriangleIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-yellow-400"><path d="m21.73 18-8-14a2 2 0 0 0-3.46 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" x2="12" y1="9" y2="13"></line><line x1="12" x2="12.01" y1="17" y2="17"></line></svg>
);

const CodeIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
);

const GitBranchIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400"><line x1="6" x2="6" y1="3" y2="15"></line><circle cx="18" cy="6" r="3"></circle><circle cx="6" cy="18" r="3"></circle><path d="M18 9a9 9 0 0 1-9 9"></path></svg>
);

const BookOpenIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2Z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7Z"></path></svg>
);

const SettingsIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6 text-gray-400"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 0 2l-.15.08a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.38a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1 0-2l.15-.08a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"></path><circle cx="12" cy="12" r="3"></circle></svg>
);

const PlusIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
);

const XIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-6 w-6"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
);

const FileIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5 text-gray-400"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>
);

const ClipboardIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path><rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect></svg>
);

const CheckIcon = () => (
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5 text-green-400"><polyline points="20 6 9 17 4 12"></polyline></svg>
);


// --- Data ---
const mockSchemaData = [
    {
        path: 'protos/eliza/v1/eliza.proto',
        content: `
syntax = "proto3";

package connectrpc.eliza.v1;

// ElizaService is a simple service that allows you to converse with Eliza.
service ElizaService {
  // Say is a unary RPC that allows you to send a sentence to Eliza and receive a response.
  rpc Say(SayRequest) returns (SayResponse) {}
  // Converse is a bidirectional RPC that allows you to stream sentences to Eliza and receive a stream of responses.
  rpc Converse(stream SayRequest) returns (stream SayResponse) {}
}

message SayRequest {
  string sentence = 1;
}

message SayResponse {
  string sentence = 1;
}
`
    },
    {
        path: 'protos/user/v1/user.proto',
        content: `
syntax = "proto3";

package user.v1;

service UserService {
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {}
    rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
}

message User {
    string id = 1;
    string name = 2;
    string email = 3;
}

message CreateUserRequest {
    string name = 1;
    string email = 2;
    string password = 3;
}

message CreateUserResponse {
    User user = 1;
}

message GetUserRequest {
    string id = 1;
}

message GetUserResponse {
    User user = 1;
}
`
    }
];

// --- Utility Functions ---

const parseSchemaForTargets = (schemaData) => {
    const targets = [];
    const packageRegex = /package\s+([\w.]+);/g;
    const serviceRegex = /service\s+(\w+)\s*{([^}]+)}/g;
    const rpcRegex = /rpc\s+(\w+)\s*\(/g;

    schemaData.forEach(file => {
        const content = file.content;
        let currentPackage = '';
        let packageMatch;
        while ((packageMatch = packageRegex.exec(content)) !== null) {
            currentPackage = packageMatch[1];
            let serviceMatch;
            serviceRegex.lastIndex = 0; 
            while ((serviceMatch = serviceRegex.exec(content)) !== null) {
                const serviceName = serviceMatch[1];
                const serviceBody = serviceMatch[2];
                let rpcMatch;
                while ((rpcMatch = rpcRegex.exec(serviceBody)) !== null) {
                    const rpcName = rpcMatch[1];
                    targets.push(`${currentPackage}.${serviceName}/${rpcName}`);
                }
            }
        }
    });
    return targets;
};

const renderJson = (json) => {
    try { return JSON.stringify(json, null, 2); } catch (e) { return "Invalid JSON"; }
};

const copyToClipboard = (text) => {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.left = "-9999px";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
        document.execCommand('copy');
    } catch (err) {
        console.error('Failed to copy text: ', err);
    }
    document.body.removeChild(textArea);
};

// --- Reusable Components ---

const StatCard = ({ icon, title, value, subValue }) => (
  <div className="bg-gray-800/50 border border-gray-700/80 rounded-lg p-5 shadow-md flex flex-col">
    <div className="flex items-center justify-between mb-2">
      <h3 className="text-sm font-medium text-gray-400">{title}</h3>
      {icon}
    </div>
    <div>
      <p className="text-3xl font-bold text-white">{value}</p>
      {subValue && <p className="text-sm text-gray-500">{subValue}</p>}
    </div>
  </div>
);

const HeaderView = ({ data }) => (
  <div className="bg-gray-800 p-3 rounded-md text-sm text-gray-300 overflow-x-auto">
    <table className="w-full">
      <tbody>
        {Object.entries(data).map(([key, value]) => (
          <tr key={key} className="border-b border-gray-700/50 last:border-b-0">
            <td className="py-1 pr-4 font-medium text-gray-400 align-top">{key}</td>
            <td className="py-1 font-mono break-all">{String(value)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);

const LogEntry = ({ log }) => {
  const [isOpen, setIsOpen] = useState(false);
  const getStatusColor = (status) => {
    if (status >= 500) return 'bg-red-500 text-red-100';
    if (status >= 400) return 'bg-yellow-500 text-yellow-100';
    if (status >= 200) return 'bg-green-500 text-green-100';
    return 'bg-gray-500 text-gray-100';
  };

  return (
    <div className="bg-gray-800/60 border border-gray-700 rounded-lg mb-3 overflow-hidden">
      <button onClick={() => setIsOpen(!isOpen)} className="w-full p-4 text-left flex items-center justify-between hover:bg-gray-700/50 transition-colors duration-200">
        <div className="flex items-center space-x-4 flex-wrap">
          <span className={`px-2 py-1 text-xs font-bold rounded-md ${getStatusColor(log.status)}`}>{log.status}</span>
          <span className="text-gray-400 font-mono text-sm">{new Date(log.timestamp).toLocaleTimeString()}</span>
          <span className="text-white font-medium">{log.service}</span>
          <span className="text-gray-400">{log.method}</span>
        </div>
        <div className="flex items-center space-x-2 flex-shrink-0 ml-4">
            <span className="text-sm text-gray-500">{log.duration}ms</span>
            <ChevronDownIcon className={`h-5 w-5 text-gray-400 transition-transform duration-300 ${isOpen ? 'rotate-180' : ''}`} />
        </div>
      </button>
      {isOpen && (
        <div className="p-4 border-t border-gray-700 bg-gray-900/50">
          <div className="space-y-8">
            <div>
              <h4 className="text-lg font-semibold text-gray-300 mb-3">Request</h4>
              <div className="space-y-4">
                  <div><h5 className="text-md font-medium text-gray-400 mb-2">Headers</h5><HeaderView data={log.requestHeaders} /></div>
                  <div><h5 className="text-md font-medium text-gray-400 mb-2">Body</h5><pre className="bg-gray-800 p-3 rounded-md text-sm text-gray-300 overflow-x-auto"><code>{renderJson(log.requestBody)}</code></pre></div>
              </div>
            </div>
            <div>
              <h4 className="text-lg font-semibold text-gray-300 mb-3">Response</h4>
              <div className="space-y-4">
                  <div><h5 className="text-md font-medium text-gray-400 mb-2">Headers</h5><HeaderView data={log.responseHeaders} /></div>
                  <div><h5 className="text-md font-medium text-gray-400 mb-2">Body</h5><pre className="bg-gray-800 p-3 rounded-md text-sm text-gray-300 overflow-x-auto"><code>{renderJson(log.responseBody)}</code></pre></div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const CommandExample = ({ title, command }) => {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        copyToClipboard(command);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div>
            <h3 className="text-lg font-semibold text-gray-300 mb-2">{title}</h3>
            <div className="relative bg-gray-900 rounded-lg p-4 font-mono text-sm text-gray-300">
                <pre className="overflow-x-auto"><code>{command}</code></pre>
                <button onClick={handleCopy} className="absolute top-3 right-3 p-2 bg-gray-700 hover:bg-gray-600 rounded-md text-gray-300 transition-colors">
                    {copied ? <CheckIcon /> : <ClipboardIcon />}
                </button>
            </div>
        </div>
    );
};

// --- Page Components ---

const SummaryPage = ({ rpcTargets }) => {
  const [stats, setStats] = useState({ totalRequests: 1234567, requestsPerSecond: 142.8, errors: 89, errorRate: '0.007%', uniqueServices: 12, uniqueMethods: 78, httpHost: 'localhost:8080', goVersion: 'go1.21.0', fauxRpcVersion: 'v0.5.2', uptime: '2d 14h 32m' });
  const [bufCurlCommand, setBufCurlCommand] = useState('');
  const [curlCommand, setCurlCommand] = useState('');
  const [isExampleExpanded, setIsExampleExpanded] = useState(false);

  useEffect(() => {
    const intervalId = setInterval(() => {
      setStats(prevStats => {
        const newRequests = prevStats.totalRequests + Math.floor(Math.random() * 15) + 5;
        const newErrors = prevStats.errors + (Math.random() < 0.02 ? 1 : 0);
        return { ...prevStats, totalRequests: newRequests, requestsPerSecond: Math.random() * 20 + 130, errors: newErrors, errorRate: (newErrors / newRequests * 100).toFixed(5) + '%' };
      });
    }, 1000);
    return () => clearInterval(intervalId);
  }, []);

  useEffect(() => {
    if (rpcTargets.length > 0) {
        const randomTarget = rpcTargets[Math.floor(Math.random() * rpcTargets.length)];
        const url = `http://${stats.httpHost}/${randomTarget}`;
        const data = `{"sentence": "hello from FauxRPC"}`;

        const newBufCurlCommand = `buf curl --schema "protos" -d '${data}' ${url}`;
        setBufCurlCommand(newBufCurlCommand);
        
        const newCurlCommand = `curl \\
    --header "Content-Type: application/json" \\
    -d '${data}' \\
    ${url}`;
        setCurlCommand(newCurlCommand);
    }
  }, [rpcTargets, stats.httpHost]);

  return (
    <div>
      <div className="bg-gray-800/50 border border-gray-700/80 rounded-lg mb-8 overflow-hidden">
        <div className="p-6">
            <h2 className="text-xl font-bold text-white mb-2">FauxRPC is running</h2>
            <p className="text-gray-400">Service is available at <code className="bg-gray-900/70 text-blue-400 px-2 py-1 rounded-md">{stats.httpHost}</code></p>
        </div>
        
        {isExampleExpanded && (
          <div className="px-6 pb-6">
            <div className="space-y-6">
                <CommandExample title="buf curl Example" command={bufCurlCommand} />
                <CommandExample title="curl Example" command={curlCommand} />
            </div>
          </div>
        )}

        <button 
            className="w-full flex justify-center items-center p-2 text-sm text-gray-400 bg-gray-900/50 hover:bg-gray-900/80 transition-colors border-t border-gray-700/80"
            onClick={() => setIsExampleExpanded(!isExampleExpanded)}
        >
            <span className="mr-2">{isExampleExpanded ? 'Hide Examples' : 'Show Examples'}</span>
            <ChevronDownIcon className={`h-5 w-5 transition-transform duration-300 ${isExampleExpanded ? 'rotate-180' : ''}`} />
        </button>
      </div>

      <h1 className="text-3xl font-bold text-white mb-6">Summary</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <StatCard icon={<ChartBarIcon />} title="Total Requests" value={stats.totalRequests.toLocaleString()} />
        <StatCard icon={<ZapIcon />} title="Requests / Sec" value={stats.requestsPerSecond.toFixed(1)} />
        <StatCard icon={<AlertTriangleIcon />} title="Errors" value={stats.errors.toLocaleString()} subValue={`${stats.errorRate} error rate`} />
        <StatCard icon={<GitBranchIcon />} title="Services" value={stats.uniqueServices.toLocaleString()} />
        <StatCard icon={<CodeIcon />} title="Methods" value={stats.uniqueMethods.toLocaleString()} />
        <StatCard icon={<ServerIcon />} title="Uptime" value={stats.uptime} />
        <StatCard icon={<CodeIcon />} title="Go Version" value={stats.goVersion} />
        <StatCard icon={<CodeIcon />} title="FauxRPC Version" value={stats.fauxRpcVersion} />
      </div>
    </div>
  );
};

const RequestLogPage = () => {
    const [logs, setLogs] = useState([]);
    const [isStreaming, setIsStreaming] = useState(true);
    const mockServices = ['UserService', 'ProductService', 'OrderService', 'PaymentService', 'InventoryService'];
    const mockMethods = { UserService: ['CreateUser', 'GetUser', 'UpdateProfile'], ProductService: ['GetProduct', 'ListProducts', 'SearchProducts'], OrderService: ['CreateOrder', 'GetOrder', 'CancelOrder'], PaymentService: ['ProcessPayment', 'GetPaymentStatus'], InventoryService: ['CheckStock', 'UpdateStock'] };
    const generateMockLog = useCallback(() => {
        const service = mockServices[Math.floor(Math.random() * mockServices.length)];
        const methods = mockMethods[service];
        const method = methods[Math.floor(Math.random() * methods.length)];
        const status = Math.random() < 0.05 ? 500 : (Math.random() < 0.1 ? 404 : 200);
        return { id: crypto.randomUUID(), timestamp: Date.now(), service, method, status, duration: (Math.random() * 200 + 10).toFixed(0), requestHeaders: { 'Content-Type': 'application/json', 'X-Request-ID': crypto.randomUUID(), 'User-Agent': 'FauxClient/1.0' }, responseHeaders: { 'Content-Type': 'application/json', 'Server': 'FauxRPC/0.5.2', 'X-Trace-ID': crypto.randomUUID() }, requestBody: { id: Math.floor(Math.random() * 1000) }, responseBody: status === 200 ? { success: true, data: `data_for_${method}` } : { success: false, error: 'An unexpected error occurred on the server.' } };
    }, []);
    useEffect(() => {
        if (isStreaming) {
            const interval = setInterval(() => { setLogs(prevLogs => [generateMockLog(), ...prevLogs.slice(0, 99)]); }, 1500);
            return () => clearInterval(interval);
        }
    }, [isStreaming, generateMockLog]);
    
    return (
        <div className="flex flex-col h-full">
            <div className="flex justify-between items-center mb-6 flex-shrink-0">
                <h1 className="text-3xl font-bold text-white">Request Log</h1>
                <div className="flex items-center space-x-4">
                    <button onClick={() => setLogs([])} className="px-4 py-2 bg-gray-600 hover:bg-gray-500 text-white font-semibold rounded-lg transition-colors">Clear</button>
                    <button onClick={() => setIsStreaming(!isStreaming)} className={`px-4 py-2 font-semibold rounded-lg transition-colors ${isStreaming ? 'bg-blue-600 hover:bg-blue-500 text-white' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'}`}>{isStreaming ? 'Pause Streaming' : 'Resume Streaming'}</button>
                </div>
            </div>
            <div className="overflow-y-auto pr-2 flex-grow">{logs.map(log => <LogEntry key={log.id} log={log} />)}</div>
        </div>
    );
};

const SchemaPage = () => {
    const [expandedFile, setExpandedFile] = useState(null);
    const handleToggleFile = (path) => { setExpandedFile(current => (current === path ? null : path)); };

    return (
        <div>
            <h1 className="text-3xl font-bold text-white mb-6">Schema</h1>
            <div className="bg-gray-800/60 border border-gray-700 rounded-lg">
                {mockSchemaData.map((file) => (
                    <div key={file.path} className={`border-b border-gray-700 last:border-b-0`}>
                        <button onClick={() => handleToggleFile(file.path)} className="w-full flex items-center justify-between p-4 text-left hover:bg-gray-700/50 transition-colors">
                            <div className="flex items-center gap-3"><FileIcon /><span className="font-mono text-gray-300">{file.path}</span></div>
                            <ChevronDownIcon className={`h-5 w-5 text-gray-400 transition-transform duration-300 ${expandedFile === file.path ? 'rotate-180' : ''}`} />
                        </button>
                        {expandedFile === file.path && (
                            <div className="p-4 border-t border-gray-700 bg-gray-900/50">
                                <pre className="text-gray-300 text-sm overflow-x-auto"><code>{file.content.trim()}</code></pre>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
};

const StubsPage = ({ rpcTargets }) => {
    const initialStubs = [{ id: "say-hello-default", target: "connectrpc.eliza.v1.ElizaService/Say", cel_content: { sentence: "req.sentence" } }];
    const [stubs, setStubs] = useState(initialStubs);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingStub, setEditingStub] = useState(null);

    const handleOpenModal = (stub = null) => { setEditingStub(stub); setIsModalOpen(true); };
    const handleCloseModal = () => { setIsModalOpen(false); setEditingStub(null); };
    const handleSaveStub = (stubData) => {
        if (editingStub) {
            setStubs(stubs.map(s => s.id === editingStub.id ? { ...stubData, id: editingStub.id } : s));
        } else {
            setStubs([...stubs, { ...stubData, id: stubData.id || crypto.randomUUID() }]);
        }
        handleCloseModal();
    };
    const handleDeleteStub = (id) => { setStubs(stubs.filter(s => s.id !== id)); };

    return (
        <div>
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold text-white">Stubs</h1>
                <button onClick={() => handleOpenModal()} className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition-colors"><PlusIcon /> Add Stub</button>
            </div>
            <div className="space-y-4">
                {stubs.map(stub => (
                    <div key={stub.id} className="bg-gray-800/60 border border-gray-700 rounded-lg p-4">
                        <div className="flex justify-between items-start">
                            <div>
                                <h3 className="text-lg font-bold text-blue-400">{stub.id}</h3>
                                <p className="text-gray-400 font-mono mt-1">{stub.target}</p>
                            </div>
                            <div className="flex space-x-2">
                                <button onClick={() => handleOpenModal(stub)} className="text-gray-400 hover:text-white">Edit</button>
                                <button onClick={() => handleDeleteStub(stub.id)} className="text-red-500 hover:text-red-400">Delete</button>
                            </div>
                        </div>
                        <pre className="bg-gray-900/70 p-3 rounded-md text-sm text-gray-300 overflow-x-auto mt-4"><code>{renderJson(stub.cel_content)}</code></pre>
                    </div>
                ))}
            </div>
            {isModalOpen && <StubModal stub={editingStub} rpcTargets={rpcTargets} onSave={handleSaveStub} onClose={handleCloseModal} />}
        </div>
    );
};

const TargetInput = ({ value, onChange, options }) => {
    const [showDropdown, setShowDropdown] = useState(false);
    const filteredOptions = options.filter(opt => opt.toLowerCase().includes(value.toLowerCase()));
    const handleSelect = (option) => { onChange(option); setShowDropdown(false); };

    return (
        <div className="relative">
            <input type="text" value={value} onChange={e => onChange(e.target.value)} onFocus={() => setShowDropdown(true)} onBlur={() => setTimeout(() => setShowDropdown(false), 150)} placeholder="e.g., connectrpc.eliza.v1.ElizaService/Say" className="w-full bg-gray-900 border border-gray-600 rounded-md px-3 py-2 text-white focus:ring-blue-500 focus:border-blue-500" required />
            {showDropdown && filteredOptions.length > 0 && (
                <div className="absolute z-10 w-full mt-1 bg-gray-700 border border-gray-600 rounded-md shadow-lg max-h-60 overflow-auto">
                    <ul>{filteredOptions.map(option => (<li key={option} onMouseDown={() => handleSelect(option)} className="px-4 py-2 text-white hover:bg-blue-600 cursor-pointer">{option}</li>))}</ul>
                </div>
            )}
        </div>
    );
};

const StubModal = ({ stub, rpcTargets, onSave, onClose }) => {
    const [id, setId] = useState(stub?.id || '');
    const [target, setTarget] = useState(stub?.target || '');
    const [celContent, setCelContent] = useState(stub ? renderJson(stub.cel_content) : '{\n  "sentence": "req.sentence"\n}');
    const [error, setError] = useState('');

    const handleSubmit = (e) => {
        e.preventDefault();
        let parsedCel;
        try { parsedCel = JSON.parse(celContent); setError(''); } catch (err) { setError('CEL Content is not valid JSON.'); return; }
        onSave({ id, target, cel_content: parsedCel });
    };

    return (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-800 rounded-lg shadow-xl p-8 w-full max-w-2xl border border-gray-700">
                <div className="flex justify-between items-center mb-6"><h2 className="text-2xl font-bold text-white">{stub ? 'Edit Stub' : 'Add Stub'}</h2><button onClick={onClose} className="text-gray-500 hover:text-white"><XIcon /></button></div>
                <form onSubmit={handleSubmit}>
                    <div className="space-y-4">
                        <div><label htmlFor="stub-id" className="block text-sm font-medium text-gray-300 mb-1">ID</label><input type="text" id="stub-id" value={id} onChange={e => setId(e.target.value)} disabled={!!stub} placeholder="e.g., say-hello-default" className="w-full bg-gray-900 border border-gray-600 rounded-md px-3 py-2 text-white focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50" /></div>
                        <div><label htmlFor="stub-target" className="block text-sm font-medium text-gray-300 mb-1">Target</label><TargetInput value={target} onChange={setTarget} options={rpcTargets} /></div>
                        <div><label htmlFor="stub-cel" className="block text-sm font-medium text-gray-300 mb-1">CEL Content (JSON)</label><textarea id="stub-cel" value={celContent} onChange={e => setCelContent(e.target.value)} rows="6" className="w-full bg-gray-900 border border-gray-600 rounded-md px-3 py-2 text-white font-mono text-sm focus:ring-blue-500 focus:border-blue-500"></textarea>{error && <p className="text-red-400 text-sm mt-1">{error}</p>}</div>
                    </div>
                    <div className="flex justify-end space-x-4 mt-8">
                        <button type="button" onClick={onClose} className="px-4 py-2 bg-gray-600 hover:bg-gray-500 text-white font-semibold rounded-lg transition-colors">Cancel</button>
                        <button type="submit" className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition-colors">Save Stub</button>
                    </div>
                </form>
            </div>
        </div>
    );
};

// --- Main App Component ---

const Sidebar = ({ currentPage, setCurrentPage }) => {
    const NavLink = ({ pageName, children }) => {
        const isActive = currentPage === pageName;
        return (<button onClick={() => setCurrentPage(pageName)} className={`flex items-center w-full px-4 py-3 rounded-lg transition-colors duration-200 ${isActive ? 'bg-blue-600 text-white' : 'text-gray-300 hover:bg-gray-700'}`}>{children}</button>);
    };
    return (
        <aside className="w-64 bg-gray-900 border-r border-gray-800 p-6 flex-col flex-shrink-0 hidden md:flex">
            <div className="flex items-center mb-10">
                <div className="bg-blue-500 p-2 rounded-lg mr-3"><ServerIcon /></div>
                <h2 className="text-2xl font-bold text-white">FauxRPC</h2>
            </div>
            <nav className="flex flex-col space-y-3">
                <NavLink pageName="summary"><ChartBarIcon /><span className="ml-4">Summary</span></NavLink>
                <NavLink pageName="request-log"><ListIcon /><span className="ml-4">Request Log</span></NavLink>
                <NavLink pageName="schema"><BookOpenIcon /><span className="ml-4">Schema</span></NavLink>
                <NavLink pageName="stubs"><SettingsIcon /><span className="ml-4">Stubs</span></NavLink>
                <a href="/fauxrpc/openapi.html" target="_blank" rel="noopener noreferrer" className="flex items-center w-full px-4 py-3 rounded-lg transition-colors duration-200 text-gray-300 hover:bg-gray-700">
                    <BookOpenIcon /><span className="ml-4">OpenAPI Docs</span>
                </a>
            </nav>
        </aside>
    );
};

export default function App() {
  const [currentPage, setCurrentPage] = useState('summary');
  const [rpcTargets, setRpcTargets] = useState([]);

  useEffect(() => {
    setRpcTargets(parseSchemaForTargets(mockSchemaData));
  }, []);

  const renderPage = () => {
    switch (currentPage) {
      case 'summary': return <SummaryPage rpcTargets={rpcTargets} />;
      case 'request-log': return <RequestLogPage />;
      case 'schema': return <SchemaPage />;
      case 'stubs': return <StubsPage rpcTargets={rpcTargets} />;
      default: return <SummaryPage rpcTargets={rpcTargets} />;
    }
  };
  return (
    <div className="bg-gray-900 text-white h-screen font-sans flex">
      <Sidebar currentPage={currentPage} setCurrentPage={setCurrentPage} />
      <main className="flex-1 p-8 overflow-auto">{renderPage()}</main>
    </div>
  );
}
