class Product < ApplicationRecord
    after_create :pubsub

    def pubsub
        msg = {
            id: self.id,
            name: self.name,
        }
        $redis.publish 'test1', msg.to_json
    end
end
