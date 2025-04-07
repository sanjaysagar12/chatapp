# **TurpleMQ**

## **What is this service?**
This service lets you send and receive messages based on topics.  
It supports two modes:
1. **Broadcast**: Real-time delivery to all active listeners.  
2. **Buffered Queue**: Messages are saved and can be retrieved later.

---

## **How to use it?**

### **1. Producer (Send Messages)**
Producers send messages to a topic with a choice of:
- **Broadcast**: Real-time delivery to active subscribers.
- **Buffered**: Save the message for later retrieval.

#### **Request Format**:
```json
{
  "role": "producer",
  "message": "Your message here",
  "topic": "Your topic here",
  "transmission_mode": "broadcast" or "buffered"
}
```

#### **Examples**:
- **Broadcast**:
  ```json
  {
    "role": "producer",
    "message": "Hello, everyone!",
    "topic": "general",
    "transmission_mode": "broadcast"
  }
  ```
- **Buffered**:
  ```json
  {
    "role": "producer",
    "message": "Important update.",
    "topic": "updates",
    "transmission_mode": "buffered"
  }
  ```

---

### **2. Consumer (Receive Messages)**
Consumers receive messages from a topic. They can:
- **Subscribe**: Receive real-time updates for a topic.
- **Pull**: Fetch saved messages from the queue.

#### **Request Format**:
```json
{
  "role": "consumer",
  "topic": "Your topic here",
  "subscribe": true (optional)
}
```

#### **Examples**:
- **Subscribe**:
  ```json
  {
    "role": "consumer",
    "subscribe": true,
    "topic": "general"
  }
  ```
- **Pull Messages**:
  ```json
  {
    "role": "consumer",
    "topic": "updates"
  }
  ```

---

### **3. Message History API**

Retrieve previously broadcasted messages for a specific topic using the REST API endpoint:

```
GET /messages?topic=<topic_name>
```

#### **Example Response**:
```json
{
  "topic": "general",
  "messages": [
    {
      "id": 42,
      "topic": "general",
      "message": "Hello, everyone!",
      "created_at": "2023-05-18T15:30:45Z"
    },
    {
      "id": 41,
      "topic": "general",
      "message": "Welcome to the channel",
      "created_at": "2023-05-18T15:25:12Z"
    }
  ]
}
```

---

## **Database Persistence**

TurpleMQ now includes SQLite database persistence for broadcast messages:

- All broadcast messages are automatically saved to a SQLite database
- Messages are stored with their topic, content, and timestamp
- Historical messages can be retrieved via the Message History API
- Database is created at `./turplemq.db` by default

---

## **Comparison: Pub-Sub vs Pull-Based**

| **Feature**              | **Publish-Subscribe**                  | **Pull-Based**                   |
|---------------------------|----------------------------------------|-----------------------------------|
| **Message Delivery**      | Push to subscribers automatically     | Consumers pull messages manually |
| **Real-Time Support**     | Yes                                   | No                               |
| **Message Storage**       | Often ephemeral; not stored long-term | Stored in a queue until consumed |
| **Consumer Control**      | Subscribers receive messages as they come | Consumers decide when to retrieve messages |
| **Scalability**           | Scales well with many subscribers     | Can scale by adding consumers    |
| **Use Cases**             | Live updates, notifications, broadcasting | Batch processing, task queues    |

---

## **Key Terms**
1. **Role**:
   - `producer`: Sends messages.
   - `consumer`: Receives messages.
2. **Topic**: The channel where messages are sent or received.
3. **Transmission Mode**:
   - `broadcast`: Real-time delivery.
   - `buffered`: Message saved for later.

---
