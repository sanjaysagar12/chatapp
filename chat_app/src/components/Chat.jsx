import { useState, useEffect, useRef } from 'react';

const Chat = () => {
  const [mode, setMode] = useState('initial'); // 'initial', 'join', 'create', 'chat'
  const [roomId, setRoomId] = useState('');
  const [username, setUsername] = useState('');
  const [message, setMessage] = useState('');
  const [messages, setMessages] = useState([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState('');
  
  const socketRef = useRef(null);
  const messagesEndRef = useRef(null);

  // Handle room creation
  const handleCreateRoom = () => {
    if (!username.trim()) {
      setError('Please enter a username');
      return;
    }
    
    // Generate a random room ID
    const randomRoomId = Math.random().toString(36).substring(2, 10);
    setRoomId(randomRoomId);
    setMode('chat');
    connectToWebSocket(randomRoomId, true);
  };

  // Handle room joining
  const handleJoinRoom = () => {
    if (!username.trim()) {
      setError('Please enter a username');
      return;
    }
    if (!roomId.trim()) {
      setError('Please enter a room ID');
      return;
    }
    
    setMode('chat');
    connectToWebSocket(roomId);
  };

  // Connect to WebSocket server
  const connectToWebSocket = (room, isNewRoom = false) => {
    const socket = new WebSocket('ws://localhost:8080');
    socketRef.current = socket;

    socket.onopen = () => {
      setConnected(true);
      console.log('WebSocket connection established');
      
      // Subscribe as consumer
      socket.send(JSON.stringify({
        role: 'consumer',
        subscribe: true,
        topic: room
      }));
      
      // Load previous messages if joining existing room
      if (!isNewRoom) {
        fetchPreviousMessages(room);
      }
    };

    socket.onmessage = (event) => {
      try {
        // Try to parse as JSON first
        const data = JSON.parse(event.data);
        let messageObj;
        
        // Check if it's a nested JSON string
        if (typeof data === 'string') {
          try {
            messageObj = JSON.parse(data);
          } catch (e) {
            messageObj = { message: data, sender: 'System', timestamp: new Date() };
          }
        } else {
          messageObj = data;
        }
        
        addMessage(username === messageObj.sender ? 'outgoing' : 'incoming', messageObj);
      } catch (e) {
        // If not JSON, add it as a text message
        addMessage('incoming', { message: event.data, sender: 'System', timestamp: new Date() });
      }
    };

    socket.onclose = () => {
      setConnected(false);
      console.log('WebSocket connection closed');
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      setError('Connection error. Please try again.');
    };
  };

  // Fetch previous messages for a room
  const fetchPreviousMessages = async (room) => {
    try {
      const response = await fetch(`http://localhost:8080/messages?topic=${room}`);
      const data = await response.json();
      
      if (data.messages && Array.isArray(data.messages)) {
        const formattedMessages = data.messages.map(msg => {
          try {
            // Try to parse the message as JSON
            const parsedMsg = JSON.parse(msg.message);
            return {
              id: msg.id,
              message: parsedMsg.message || msg.message,
              sender: parsedMsg.sender || 'Unknown User',
              timestamp: new Date(msg.created_at),
              type: 'history'
            };
          } catch (e) {
            // If not valid JSON, use the raw message
            return {
              id: msg.id,
              message: msg.message,
              sender: 'Previous User',
              timestamp: new Date(msg.created_at),
              type: 'history'
            };
          }
        });
        
        setMessages(formattedMessages.reverse());
      }
    } catch (error) {
      console.error('Error fetching messages:', error);
      setError('Failed to load previous messages');
    }
  };

  // Send a message
  const sendMessage = () => {
    if (!message.trim() || !connected) return;

    const messageObj = {
      sender: username,
      message: message,
      timestamp: new Date()
    };
    
    socketRef.current.send(JSON.stringify({
      role: 'producer',
      message: JSON.stringify(messageObj),
      topic: roomId,
      transmission_mode: 'broadcast'
    }));
    
    setMessage('');
  };

  // Add a message to the chat
  const addMessage = (type, messageData) => {
    setMessages(prevMessages => [...prevMessages, { ...messageData, type }]);
  };

  // Format date for display
  const formatMessageTime = (timestamp) => {
    if (!timestamp) return '';
    
    let date;
    if (typeof timestamp === 'string') {
      date = new Date(timestamp);
    } else {
      date = timestamp;
    }
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      return '';
    }
    
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  // Scroll to bottom of messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Clean up WebSocket connection on unmount
  useEffect(() => {
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, []);

  // Handle Enter key press
  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      sendMessage();
    }
  };

  return (
    <div className="flex flex-col h-screen bg-gray-100">
      {mode === 'initial' && (
        <div className="flex flex-col items-center justify-center h-full p-4">
          <div className="w-full max-w-md bg-white rounded-lg shadow-md p-6">
            <h2 className="text-2xl font-bold text-center text-gray-800 mb-6">Welcome to TurpleMQ Chat</h2>
            
            {error && <div className="p-3 mb-4 text-sm text-red-700 bg-red-100 rounded">{error}</div>}
            
            <div className="mb-4">
              <label className="block text-gray-700 text-sm font-bold mb-2">Enter your username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                placeholder="Username"
              />
            </div>
            
            <div className="flex flex-col space-y-4">
              <button
                onClick={() => setMode('join')}
                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Join a Room
              </button>
              
              <button
                onClick={() => setMode('create')}
                className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Create a Room
              </button>
            </div>
          </div>
        </div>
      )}

      {mode === 'join' && (
        <div className="flex flex-col items-center justify-center h-full p-4">
          <div className="w-full max-w-md bg-white rounded-lg shadow-md p-6">
            <h2 className="text-2xl font-bold text-center text-gray-800 mb-6">Join a Chat Room</h2>
            
            {error && <div className="p-3 mb-4 text-sm text-red-700 bg-red-100 rounded">{error}</div>}
            
            <div className="mb-4">
              <label className="block text-gray-700 text-sm font-bold mb-2">Room ID</label>
              <input
                type="text"
                value={roomId}
                onChange={(e) => setRoomId(e.target.value)}
                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                placeholder="Enter room ID"
              />
            </div>
            
            <div className="flex justify-between">
              <button
                onClick={() => setMode('initial')}
                className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Back
              </button>
              
              <button
                onClick={handleJoinRoom}
                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Join Room
              </button>
            </div>
          </div>
        </div>
      )}

      {mode === 'create' && (
        <div className="flex flex-col items-center justify-center h-full p-4">
          <div className="w-full max-w-md bg-white rounded-lg shadow-md p-6">
            <h2 className="text-2xl font-bold text-center text-gray-800 mb-6">Create a Chat Room</h2>
            
            {error && <div className="p-3 mb-4 text-sm text-red-700 bg-red-100 rounded">{error}</div>}
            
            <p className="text-gray-600 mb-4">
              Create a new room and share the Room ID with others to invite them.
            </p>
            
            <div className="flex justify-between">
              <button
                onClick={() => setMode('initial')}
                className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Back
              </button>
              
              <button
                onClick={handleCreateRoom}
                className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Create Room
              </button>
            </div>
          </div>
        </div>
      )}

      {mode === 'chat' && (
        <div className="flex flex-col h-full">
          <div className="bg-blue-600 text-white p-4 shadow-md">
            <div className="flex justify-between items-center">
              <div>
                <h2 className="text-xl font-bold">Room: {roomId}</h2>
                <p className="text-sm">Logged in as: {username}</p>
              </div>
              <div>
                <span className={`inline-block w-3 h-3 rounded-full ${connected ? 'bg-green-400' : 'bg-red-500'} mr-2`}></span>
                <span>{connected ? 'Connected' : 'Disconnected'}</span>
              </div>
            </div>
          </div>
          
          <div className="flex-grow overflow-y-auto p-4 space-y-4">
            {messages.map((msg, index) => (
              <div 
                key={index} 
                className={`flex ${msg.type === 'outgoing' ? 'justify-end' : 'justify-start'}`}
              >
                <div 
                  className={`max-w-xs md:max-w-md rounded-lg px-4 py-2 ${
                    msg.type === 'outgoing' 
                      ? 'bg-blue-500 text-white' 
                      : msg.type === 'history' 
                        ? 'bg-gray-200 text-gray-800' 
                        : 'bg-gray-300 text-gray-800'
                  }`}
                >
                  <div className="font-bold text-sm">{msg.sender}</div>
                  <div>{typeof msg.message === 'string' ? msg.message : JSON.stringify(msg.message)}</div>
                  <div className="text-xs opacity-75 text-right">
                    {formatMessageTime(msg.timestamp)}
                  </div>
                </div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>
          
          {error && <div className="p-3 mb-4 text-sm text-red-700 bg-red-100 rounded">{error}</div>}
          
          <div className="bg-gray-200 p-4">
            <div className="flex">
              <input
                type="text"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                className="flex-grow shadow appearance-none border rounded py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                placeholder="Type your message..."
                disabled={!connected}
              />
              <button
                onClick={sendMessage}
                disabled={!connected}
                className="ml-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline disabled:bg-gray-400"
              >
                Send
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Chat;