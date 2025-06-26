import json
from router import route_action

def handler(event, context):
    try:
        body = json.loads(event['body'])
        return route_action(body)
    except Exception as e:
        return {"statusCode": 500, "body": f"Internal error: {str(e)}"}

if __name__ == "__main__":
    with open("event.json") as f:
        event = json.load(f)
    result = handler(event, None)
    print(json.dumps(result, indent=2))