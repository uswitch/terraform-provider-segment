# An Amazon Kinesis destination
resource "segment_destination" "test__amazon-kinesis" {
  source          = "simple_test"
  name            = "amazon-kinesis"
  enabled         = true
  connection_mode = "UNSPECIFIED"

  # This part will vary depending on the destination.
  config = {
    "region" = jsonencode(
      {
        type  = "string"
        value = "eu-west-1"
      }
    )
    "roleAddress" = jsonencode(
      {
        type  = "string"
        value = "arn:aws:iam::123456789:role/segment_kinesis_role"
      }
    )
    "secretId" = jsonencode(
      {
        type  = "string"
        value = "ABC-123"
      }
    )
    "stream" = jsonencode(
      {
        type  = "string"
        value = "segment-event-stream"
      }
    )
    "useMessageId" = jsonencode(
      {
        type  = "boolean"
        value = false
      }
    )
  }
}
