import json
from router import route_action

def handler(event, context):
    try:
        body = json.loads(event.get('body', '{}'))
        res = route_action(body)
        return {
            "statusCode": res.get("status_code", 500),
            "body": json.dumps(res.get("json", {}))
        }
        
    except Exception as e:
        return {
            "statusCode": 500,
            "body": json.dumps({"error": f"Internal error: {str(e)}"})
        }

if __name__ == "__main__":
    with open("event.json") as f:
        event = json.load(f)
    result = handler(event, None)
    print(json.dumps(result, indent=2))
